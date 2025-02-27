package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/densestvoid/krogerrecipeshopper/assets"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Domain       string
}

func New(ctx context.Context, logger *slog.Logger, config Config, repo *data.Repository, cache *data.Cache) http.Handler {
	mux := chi.NewRouter()
	mux.Use(
		httprate.LimitByIP(100, time.Minute),
		NewErrorStatusMiddleware(),
		NewSlogMiddleware(),
	)

	// Handlers for favicons
	mux.Handle("/*", middleware.SetHeader("Cache-Control", "max-age=86400")(
		http.FileServer(http.FS(assets.Files)),
	))

	mux.Route("/auth", NewAuthMux(config, repo))
	mux.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware(config, repo))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			account, err := repo.GetAccountByID(ctx, authCookies.AccountID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			switch account.Homepage {
			case data.HomepageOptionRecipes:
				http.Redirect(w, r, "/recipes", http.StatusTemporaryRedirect)
			case data.HomepageOptionFavorites:
				http.Redirect(w, r, "/recipes/favorites", http.StatusTemporaryRedirect)
			case data.HomepageOptionExplore:
				http.Redirect(w, r, "/recipes/explore", http.StatusTemporaryRedirect)
			default:
				http.Redirect(w, r, "/welcome", http.StatusTemporaryRedirect)
			}
		})
		r.Get("/welcome", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.WelcomePage().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Route("/accounts", NewAccountMux(config, repo, cache))
		r.Route("/profiles", NewProfilesMux(repo))
		r.Route("/recipes", NewRecipesMux(config, repo, cache))
		r.Route("/products", NewProductsMux(config, repo, cache))
		r.Route("/locations", NewLocationsMux(config))
		r.Route("/cart", NewCartMux(config, repo, cache))
		r.Route("/shopping-list", NewShoppingListMux(config, repo, cache))
	})

	return mux
}
