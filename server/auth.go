package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
			account, err := repo.GetAccountByKrogerProfileID(r.Context(), profileResp.Profile.ID)
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
			session, err := repo.CreateSession(r.Context(), account.ID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create session: %v", err), http.StatusInternalServerError)
				return
			}
			if err := SetAuthResponseCookies(r.Context(), w, session, authResp); err != nil {
				http.Error(w, fmt.Sprintf("Unable to set auth cookies: %v", err), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusFound)
		})

		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			ClearAuthCookies(w)
		})
	}
}

func (c Config) RedirectUrl() string {
	return fmt.Sprintf("%s/auth", c.Domain)
}

func LoginRedirectURL(config Config, scopes ...string) string {
	scopesURIEncoded := url.QueryEscape(strings.Join(scopes, " "))
	return fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		kroger.OAuth2BaseURL,
		config.ClientID,
		config.RedirectUrl(),
		scopesURIEncoded,
	)
}

type ContextAuthCookies struct{}

type AuthCookies struct {
	AccessToken  string
	RefreshToken string
	SessionID    uuid.UUID
	AccountID    uuid.UUID
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request, url string, err error) {
	slog.Error("failed to authenticate response", "error", err)
	if r.Header.Get("HX-Request") != "" {
		w.Header().Add("HX-Redirect", url)
	} else {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func AuthenticationMiddleware(config Config, repo *data.Repository) func(next http.Handler) http.Handler {
	loginRedirectURL := LoginRedirectURL(config, kroger.ScopeCartBasicWrite, kroger.ScopeProfileCompact)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			refreshToken := ""
			refreshTokenCookie, err := r.Cookie("refreshToken")
			if err != nil || refreshTokenCookie.Value == "" || refreshTokenCookie.Valid() != nil {
				RedirectToLogin(w, r, loginRedirectURL, errors.New("refreshToken missing"))
				return
			}
			refreshToken = refreshTokenCookie.Value

			accessToken := ""
			sessionID := uuid.Nil
			accountID := uuid.Nil
			accessTokenCookie, err := r.Cookie("accessToken")
			if err != nil || accessTokenCookie.Value == "" || accessTokenCookie.Valid() != nil {
				authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
				authResp, err := authClient.PostToken(r.Context(), kroger.RefreshToken{
					RefreshToken: refreshToken,
				})
				if err != nil {
					RedirectToLogin(w, r, loginRedirectURL, err)
					return
				}
				identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
				profileResp, err := identityClient.GetProfile(r.Context())
				if err != nil {
					RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to get kroger profile id: %w", err))
					return
				}
				account, err := repo.GetAccountByKrogerProfileID(r.Context(), profileResp.Profile.ID)
				// Refrsh token exists, the user has logged in before and should have an account already
				if err != nil {
					RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to get account: %w", err))
					return
				}
				session, err := repo.CreateSession(r.Context(), account.ID)
				if err != nil {
					RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to create session: %w", err))
					return
				}
				if err := SetAuthResponseCookies(r.Context(), w, session, authResp); err != nil {
					RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to set auth cookies: %w", err))
					return
				}

				refreshToken = authResp.RefreshToken
				accessToken = authResp.AccessToken
				sessionID = session.ID
				accountID = account.ID
			} else {
				accessToken = accessTokenCookie.Value
			}

			// Session not created when getting new access token
			if sessionID == uuid.Nil {
				// Session ID cookie exists
				if sessionIDCookie, err := r.Cookie("sessionID"); err == nil && sessionIDCookie.Value != "" && sessionIDCookie.Valid() == nil {
					sessionID, err = uuid.Parse(sessionIDCookie.Value)
					if err != nil {
						RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to parse session id: %w", err))
						return
					}
					session, err := repo.GetSessionByID(r.Context(), sessionID)
					if err != nil {
						RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to create session: %w", err))
						return
					}
					accountID = session.AccountID
				} else { // Session ID cookie missing
					identityClient := kroger.NewIdentityClient(http.DefaultClient, kroger.PublicEnvironment, accessToken)
					profileResp, err := identityClient.GetProfile(r.Context())
					if err != nil {
						RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to get kroger profile id: %w", err))
						return
					}
					account, err := repo.GetAccountByKrogerProfileID(r.Context(), profileResp.Profile.ID)
					// Refrsh token exists, the user has logged in before and should have an account already
					if err != nil {
						RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to get account: %w", err))
						return
					}
					session, err := repo.CreateSession(r.Context(), account.ID)
					if err != nil {
						RedirectToLogin(w, r, loginRedirectURL, fmt.Errorf("unable to create session: %w", err))
						return
					}
					sessionID = session.ID
					accountID = account.ID

					http.SetCookie(w, &http.Cookie{
						Path:     "/",
						Name:     "sessionID",
						Value:    session.ID.String(),
						Secure:   true,
						HttpOnly: true,
						SameSite: http.SameSiteLaxMode,
					})
				}
			}

			ctx := context.WithValue(r.Context(), ContextAuthCookies{}, AuthCookies{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				SessionID:    sessionID,
				AccountID:    accountID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

const RefreshTokenMaxAgeSeconds = 15_768_000 // 6 months in seconds, taken from an FAQ response

func SetAuthResponseCookies(ctx context.Context, w http.ResponseWriter, session data.Session, credentials *kroger.PostTokenResponse) error {
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
		MaxAge:   RefreshTokenMaxAgeSeconds,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "sessionID",
		Value:    session.ID.String(),
		MaxAge:   credentials.ExpiresIn,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func ClearAuthCookies(w http.ResponseWriter) {
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
		Name:     "sessionID",
		Value:    "",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func GetAuthCookies(r *http.Request) (AuthCookies, error) {
	authCookies, ok := r.Context().Value(ContextAuthCookies{}).(AuthCookies)
	if !ok {
		return AuthCookies{}, fmt.Errorf("auth cookies missing")
	}
	return authCookies, nil
}
