package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewAuthMux(config Config, repo *data.Repository) func(r chi.Router) {
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
			identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
			profileResp, err := identityClient.GetProfile(r.Context())
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to get kroger profile id: %v", err), http.StatusInternalServerError)
				return
			}
			account, err := repo.GetAccountID(r.Context(), profileResp.Profile.ID)
			if errors.Is(err, sql.ErrNoRows) {
				// This is the user's first time logging in
				if account, err = repo.CreateAccount(r.Context(), profileResp.Profile.ID); err != nil {
					http.Error(w, fmt.Sprintf("Unable to create account: %v", err), http.StatusInternalServerError)
					return
				}
			} else if err != nil {
				http.Error(w, fmt.Sprintf("Unable to get account: %v", err), http.StatusInternalServerError)
				return
			}
			SetAuthResponseCookies(r.Context(), w, account, authResp)
			http.Redirect(w, r, "/", http.StatusFound)
		})

		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Path:     "/",
				Name:     "accessToken",
				Value:    "",
				Expires:  time.Now(),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.SetCookie(w, &http.Cookie{
				Path:     "/",
				Name:     "refreshToken",
				Value:    "",
				Expires:  time.Now(),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.SetCookie(w, &http.Cookie{
				Path:     "/",
				Name:     "accountID",
				Value:    "",
				Expires:  time.Now(),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
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

func AuthenticationMiddleware(config Config, repo *data.Repository) func(next http.Handler) http.Handler {
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
				identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
				profileResp, err := identityClient.GetProfile(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf("Unable to get kroger profile id: %v", err), http.StatusInternalServerError)
					return
				}
				account, err := repo.GetAccountID(r.Context(), profileResp.Profile.ID)
				// Refrsh token exists, the user has logged in before and should have an account already
				if err != nil {
					http.Error(w, fmt.Sprintf("Unable to get account: %v", err), http.StatusInternalServerError)
					return
				}
				SetAuthResponseCookies(r.Context(), w, account, authResp)
				http.Redirect(w, r, r.RequestURI, http.StatusTemporaryRedirect)
				return
			}
			loginRedirect(w, r)
		})
	}
}

func SetAuthResponseCookies(ctx context.Context, w http.ResponseWriter, account data.Account, credentials *kroger.PostTokenResponse) error {
	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "accessToken",
		Value:    credentials.AccessToken,
		MaxAge:   credentials.ExpiresIn,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "refreshToken",
		Value:    credentials.RefreshToken,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "accountID",
		Value:    account.ID.String(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func GetAccountIDRequestCookie(r *http.Request) (uuid.UUID, error) {
	var accountID uuid.UUID
	if accountIDCookie, err := r.Cookie("accountID"); err != nil {
		return uuid.Nil, fmt.Errorf("Account ID cookie not found: %w", err)
	} else if err = accountIDCookie.Valid(); err != nil {
		return uuid.Nil, fmt.Errorf("Invalid Account ID cookie: %w", err)
	} else if accountIDCookie.Value == "" {
		return uuid.Nil, fmt.Errorf("Account ID cookie empty")
	} else if accountID, err = uuid.Parse(accountIDCookie.Value); err != nil {
		return uuid.Nil, fmt.Errorf("Invalid Account ID: %w", err)
	}
	return accountID, nil
}
