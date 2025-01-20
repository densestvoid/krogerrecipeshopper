package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewAuthMux(config Config) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
			authResp, err := authClient.PostToken(r.Context(), kroger.AuthorizationCode{
				Code:        r.FormValue("code"),
				RedirectURI: config.RedirectUrl(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			SetAuthResponseCookies(r.Context(), w, authResp)
			http.Redirect(w, r, "/", http.StatusFound)
		})

		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "accessToken",
				Value:   "",
				Expires: time.Now(),
				Secure:  true,
			})
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "refreshToken",
				Value:   "",
				Expires: time.Now(),
				Secure:  true,
			})
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "userID",
				Value:   "",
				Expires: time.Now(),
				Secure:  true,
			})
			w.Header().Add("HX-Refresh", "true")
			w.WriteHeader(http.StatusOK)
		})
	}
}

func (c Config) RedirectUrl() string {
	return fmt.Sprintf("%s/auth", c.Domain)
}

func LoginRedirect(config Config, scopes ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		scopesURIEncoded := url.QueryEscape(strings.Join(scopes, " "))
		redirectURL := fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			kroger.OAuth2BaseURL,
			config.ClientID,
			config.RedirectUrl(),
			scopesURIEncoded,
		)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

func AuthenticationMiddleware(config Config) func(next http.Handler) http.Handler {
	loginRedirect := LoginRedirect(config, kroger.ScopeCartBasicWrite, kroger.ScopeProfileCompact)
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
					loginRedirect(w, r)
					return
				}
				SetAuthResponseCookies(r.Context(), w, authResp)
				http.Redirect(w, r, r.RequestURI, http.StatusTemporaryRedirect)
				return
			}
			loginRedirect(w, r)
		})
	}
}

func SetAuthResponseCookies(ctx context.Context, w http.ResponseWriter, credentials *kroger.PostTokenResponse) error {
	identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, credentials.AccessToken)
	profileResp, err := identityClient.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("Unable to get user id: %w", err)
	}
	http.SetCookie(w, &http.Cookie{
		Path:   "/",
		Name:   "accessToken",
		Value:  credentials.AccessToken,
		MaxAge: credentials.ExpiresIn,
		Secure: true,
	})
	http.SetCookie(w, &http.Cookie{
		Path:   "/",
		Name:   "refreshToken",
		Value:  credentials.RefreshToken,
		Secure: true,
	})
	http.SetCookie(w, &http.Cookie{
		Path:   "/",
		Name:   "userID",
		Value:  profileResp.Profile.ID.String(),
		Secure: true,
	})
	return nil
}

func GetUserIDRequestCookie(r *http.Request) (uuid.UUID, error) {
	var userID uuid.UUID
	if userIDCookie, err := r.Cookie("userID"); err != nil {
		return uuid.Nil, fmt.Errorf("User ID cookie not found: %w", err)
	} else if err = userIDCookie.Valid(); err != nil {
		return uuid.Nil, fmt.Errorf("Invalid User ID cookie: %w", err)
	} else if userIDCookie.Value == "" {
		return uuid.Nil, fmt.Errorf("User ID cookie empty")
	} else if userID, err = uuid.Parse(userIDCookie.Value); err != nil {
		return uuid.Nil, fmt.Errorf("Invalid User ID: %w", err)
	}
	return userID, nil
}
