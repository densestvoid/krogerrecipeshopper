package server

import (
	"fmt"
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewAccountMux(repo *data.Repository) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
				return
			}

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

			if err := templates.Account(account, profile).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		r.Get("/profiles", func(w http.ResponseWriter, r *http.Request) {
			profiles, err := repo.ListProfiles(r.Context())
			if err != nil {
				http.Error(w, fmt.Sprintf("getting profiles: %v", err), http.StatusInternalServerError)
				return
			}

			if err := templates.Profiles(profiles).Render(w); err != nil {
				http.Error(w, fmt.Sprintf("writing profiles page: %v", err), http.StatusInternalServerError)
				return
			}
		})

		r.Route("/{accountID}", func(r chi.Router) {
			r.Patch("/settings", func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()
				accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
				imageSize := r.FormValue("imageSize")
				if err := repo.UpdateAccountImageSize(r.Context(), accountID, imageSize); err != nil {
					http.Error(w, fmt.Sprintf("updating account image size: %v", err), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Route("/profile", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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

					if err := templates.Profile(account, profile).Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					r.ParseForm()
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					displayName := r.FormValue("displayName")
					if _, err := repo.CreateProfile(r.Context(), accountID, displayName); err != nil {
						http.Error(w, fmt.Sprintf("creating profile: %v", err), http.StatusInternalServerError)
						return
					}
					w.Header().Add("HX-Trigger", "profile-update")
				})

				r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
					r.ParseForm()
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					displayName := r.FormValue("displayName")
					if err := repo.UpdateProfileDisplayName(r.Context(), accountID, displayName); err != nil {
						http.Error(w, fmt.Sprintf("updating profile: %v", err), http.StatusInternalServerError)
						return
					}

					w.Header().Add("HX-Trigger", "profile-update")
				})

				r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if err := repo.DeleteProfile(r.Context(), accountID); err != nil {
						http.Error(w, fmt.Sprintf("deleting profile: %v", err), http.StatusInternalServerError)
						return
					}
					w.Header().Add("HX-Trigger", "profile-update")
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
		})
	}
}
