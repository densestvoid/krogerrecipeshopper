package recipes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates/pages"
)

type Config struct {
	ClientID      string
	ClientSecret  string
	OAuth2BaseURL string
	RedirectUrl   string
}

func LoginRedirect(config Config, scope string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			config.OAuth2BaseURL,
			config.ClientID,
			config.RedirectUrl,
			scope,
		)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

func AuthenticationMiddleware(config Config) func(next http.Handler) http.Handler {
	loginRedirect := LoginRedirect(config, kroger.ScopeCartBasicWrite)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, err := r.Cookie("accessToken"); err == nil && token.Value != "" && token.Valid() == nil {
				next.ServeHTTP(w, r)
				return
			} else if token, err := r.Cookie("refreshToken"); err == nil && token.Value != "" && token.Valid() == nil {
				authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
				authResp, err := authClient.PostToken(r.Context(), kroger.RefreshToken{
					RefreshToken: token.Value,
				})
				if err != nil {
					http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
					return
				}
				identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
				profileResp, err := identityClient.GetProfile(r.Context())
				if err != nil {
					http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:   "accessToken",
					Value:  authResp.AccessToken,
					MaxAge: authResp.ExpiresIn,
				})
				http.SetCookie(w, &http.Cookie{
					Name:  "refreshToken",
					Value: authResp.RefreshToken,
				})
				http.SetCookie(w, &http.Cookie{
					Name:  "userID",
					Value: profileResp.Profile.ID.String(),
				})
				http.Redirect(w, r, r.RequestURI, http.StatusTemporaryRedirect)
				return
			}
			loginRedirect(w, r)
		})
	}
}

func NewServer(ctx context.Context, logger *slog.Logger, config Config, repo *data.Repository) http.Handler {
	mux := chi.NewRouter()
	chiLogger := httplog.NewLogger("Kroger Recipes", httplog.Options{
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		RequestHeaders:  true,
		ResponseHeaders: true,
	})
	mux.Use(httplog.RequestLogger(chiLogger))

	mux.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
		authResp, err := authClient.PostToken(r.Context(), kroger.AuthorizationCode{
			Code:        r.FormValue("code"),
			RedirectURI: config.RedirectUrl,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
		profileResp, err := identityClient.GetProfile(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "accessToken",
			Value:  authResp.AccessToken,
			MaxAge: authResp.ExpiresIn,
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "refreshToken",
			Value: authResp.RefreshToken,
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "userID",
			Value: profileResp.Profile.ID.String(),
		})
		http.Redirect(w, r, "/", http.StatusFound)
	})

	mux.Route("/", func(r chi.Router) {
		r.Use(AuthenticationMiddleware(config))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := pages.DashboardPage().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Get("/recipes", func(w http.ResponseWriter, r *http.Request) {
			if err := pages.Recipes().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Get("/recipes/table", func(w http.ResponseWriter, r *http.Request) {
			recipes, err := repo.ListRecipes(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := pages.RecipeTable(recipes).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Post("/recipes", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()

			name := r.FormValue("name")
			if name == "" {
				http.Error(w, "Name is required", http.StatusBadRequest)
				return
			}

			description := r.FormValue("description")
			if description == "" {
				http.Error(w, "Description is required", http.StatusBadRequest)
				return
			}

			if idString := r.FormValue("id"); idString != "" {
				id, err := uuid.Parse(idString)
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				if err := repo.UpdateRecipe(r.Context(), data.Recipe{
					ID:          id,
					UserID:      uuid.Nil,
					Name:        name,
					Description: description,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				_, err := repo.CreateRecipe(r.Context(), uuid.Nil, name, description)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Add("HX-Trigger", "recipe-update")
		})
		r.Get("/recipes/modal", func(w http.ResponseWriter, r *http.Request) {
			if err := pages.RecipeDetailsForm(nil).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Get("/recipes/modal/{recipeID}", func(w http.ResponseWriter, r *http.Request) {
			recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))
			recipe, err := repo.GetRecipe(r.Context(), recipeID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := pages.RecipeDetailsForm(&recipe).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Delete("/recipes/{recipeID}", func(w http.ResponseWriter, r *http.Request) {
			recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))
			if err := repo.DeleteRecipe(r.Context(), recipeID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Add("HX-Trigger", "recipe-update")
		})
	})

	return mux
}
