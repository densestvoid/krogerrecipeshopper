package server

import (
	"fmt"
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/app"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewAccountMux(config Config, repo *data.Repository, cache *data.Cache) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			account, err := repo.GetAccountByID(r.Context(), authCookies.AccountID)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting account: %v", err), http.StatusInternalServerError)
				return
			}

			profile, err := repo.GetProfileByAccountID(r.Context(), authCookies.AccountID)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting profile: %v", err), http.StatusInternalServerError)
				return
			}

			authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
			authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{
				Scope: kroger.ScopeProductCompact,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
			locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

			krogerManager := app.NewKrogerManager(productsClient, locationsClient, cache)

			var location *data.CacheLocation
			if account.LocationID != nil {
				loc, err := krogerManager.GetLocation(r.Context(), *account.LocationID)
				if err != nil {
					http.Error(w, fmt.Sprintf("getting location: %v", err), http.StatusInternalServerError)
					return
				}
				location = &loc
			}

			if err := templates.AccountPage(templates.Account{
				ID:        account.ID,
				ImageSize: account.ImageSize,
				Location:  location,
				Homepage:  account.Homepage,
			}, profile).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/profiles", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.Profiles().Render(w); err != nil {
				http.Error(w, fmt.Sprintf("writing profiles page: %v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/profiles/search", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			name := r.FormValue("name")
			if name == "" {
				w.WriteHeader(http.StatusOK)
				return
			}

			profiles, err := repo.ListProfiles(r.Context(), name)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting profiles: %v", err), http.StatusInternalServerError)
				return
			}

			if err := templates.ProfilesSearchResults(profiles).Render(w); err != nil {
				http.Error(w, fmt.Sprintf("writing profiles search results: %v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Route("/{accountID}", func(r chi.Router) {
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
				if authCookies.AccountID != requestedAccountID {
					http.Error(w, "deleting account", http.StatusUnauthorized)
					return
				}
				if err := repo.DeleteAccount(r.Context(), requestedAccountID); err != nil {
					http.Error(w, fmt.Sprintf("deleting account: %v", err), http.StatusInternalServerError)
					return
				}
				ClearAuthCookies(w)
			})

			r.Patch("/settings", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
				if authCookies.AccountID != requestedAccountID {
					http.Error(w, "deleting account", http.StatusUnauthorized)
					return
				}

				if err := r.ParseForm(); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

				if r.Form.Has("imageSize") {
					imageSize := r.FormValue("imageSize")
					if err := repo.UpdateAccountImageSize(r.Context(), accountID, imageSize); err != nil {
						http.Error(w, fmt.Sprintf("updating account image size: %v", err), http.StatusInternalServerError)
						return
					}
				}

				if r.Form.Has("homepage") {
					homepage := r.FormValue("homepage")
					if err := repo.UpdateAccountHomepage(r.Context(), accountID, homepage); err != nil {
						http.Error(w, fmt.Sprintf("updating account homepage: %v", err), http.StatusInternalServerError)
						return
					}
				}

				if r.Form.Has("locationID") {
					locationID := r.FormValue("locationID")
					var locationIDStr *string
					if locationID != "" {
						locationIDStr = &locationID
					}
					if err := repo.UpdateAccountLocationID(r.Context(), accountID, locationIDStr); err != nil {
						http.Error(w, fmt.Sprintf("updating account location ID: %v", err), http.StatusInternalServerError)
						return
					}

					account, err := repo.GetAccountByID(r.Context(), accountID)
					if err != nil {
						http.Error(w, fmt.Sprintf("getting account: %v", err), http.StatusInternalServerError)
						return
					}

					authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
					authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{
						Scope: kroger.ScopeProductCompact,
					})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
					locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

					krogerManager := app.NewKrogerManager(productsClient, locationsClient, cache)

					var location *data.CacheLocation
					if account.LocationID != nil {
						loc, err := krogerManager.GetLocation(r.Context(), *account.LocationID)
						if err != nil {
							http.Error(w, fmt.Sprintf("getting location: %v", err), http.StatusInternalServerError)
							return
						}
						location = &loc
					}

					if err := templates.LocationNode(location).Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				w.WriteHeader(http.StatusOK)
			})

			r.Route("/profile", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					authCookies, err := GetAuthCookies(r)
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
						return
					}
					requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if authCookies.AccountID != requestedAccountID {
						http.Error(w, "deleting account", http.StatusUnauthorized)
						return
					}

					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

					account, err := repo.GetAccountByID(r.Context(), accountID)
					if err != nil {
						http.Error(w, fmt.Sprintf("getting account: %v", err), http.StatusInternalServerError)
						return
					}

					profile, err := repo.GetProfileByAccountID(r.Context(), accountID)
					if err != nil {
						http.Error(w, fmt.Sprintf("getting profile: %v", err), http.StatusInternalServerError)
						return
					}

					authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
					authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{
						Scope: kroger.ScopeProductCompact,
					})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
					locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

					krogerManager := app.NewKrogerManager(productsClient, locationsClient, cache)

					var location *data.CacheLocation
					if account.LocationID != nil {
						loc, err := krogerManager.GetLocation(r.Context(), *account.LocationID)
						if err != nil {
							http.Error(w, fmt.Sprintf("getting location: %v", err), http.StatusInternalServerError)
							return
						}
						location = &loc
					}

					if err := templates.Profile(templates.Account{
						ID:        account.ID,
						ImageSize: account.ImageSize,
						Location:  location,
						Homepage:  account.Homepage,
					}, profile).Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					authCookies, err := GetAuthCookies(r)
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
						return
					}
					requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if authCookies.AccountID != requestedAccountID {
						http.Error(w, "deleting account", http.StatusUnauthorized)
						return
					}

					if err := r.ParseForm(); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					displayName := r.FormValue("displayName")

					if _, err := repo.CreateProfile(r.Context(), accountID, displayName); err != nil {
						http.Error(w, fmt.Sprintf("creating profile: %v", err), http.StatusInternalServerError)
						return
					}
					w.Header().Add("HX-Trigger", "profile-update")
					w.WriteHeader(http.StatusOK)
				})

				r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
					authCookies, err := GetAuthCookies(r)
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
						return
					}
					requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if authCookies.AccountID != requestedAccountID {
						http.Error(w, "deleting account", http.StatusUnauthorized)
						return
					}

					if err := r.ParseForm(); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					displayName := r.FormValue("displayName")

					if err := repo.UpdateProfileDisplayName(r.Context(), accountID, displayName); err != nil {
						http.Error(w, fmt.Sprintf("updating profile: %v", err), http.StatusInternalServerError)
						return
					}

					w.Header().Add("HX-Trigger", "profile-update")
					w.WriteHeader(http.StatusOK)
				})

				r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
					authCookies, err := GetAuthCookies(r)
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
						return
					}
					requestedAccountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if authCookies.AccountID != requestedAccountID {
						http.Error(w, "deleting account", http.StatusUnauthorized)
						return
					}

					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if err := repo.DeleteProfile(r.Context(), accountID); err != nil {
						http.Error(w, fmt.Sprintf("deleting profile: %v", err), http.StatusInternalServerError)
						return
					}
					w.Header().Add("HX-Trigger", "profile-update")
					w.WriteHeader(http.StatusOK)
				})
			})
		})
	}
}

func NewProfilesMux(repo *data.Repository) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/{accountID}", func(w http.ResponseWriter, r *http.Request) {
			accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

			profile, err := repo.GetProfileByAccountID(r.Context(), accountID)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting profile: %v", err), http.StatusInternalServerError)
				return
			} else if profile == nil {
				http.Error(w, "No profile found for this account", http.StatusInternalServerError)
				return
			}

			if err := templates.ProfilePage(*profile).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
	}
}
