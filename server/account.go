package server

import (
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewAccountMux(repo *data.Repository) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDRequestCookie(r)
			if err != nil {
				Error(w, "Getting account id", err, http.StatusUnauthorized)
				return
			}

			account, err := repo.GetAccountByID(r.Context(), accountID)
			if err != nil {
				Error(w, "Getting account", err, http.StatusInternalServerError)
				return
			}

			profile, err := repo.GetProfileByAccountID(r.Context(), accountID)
			if err != nil {
				Error(w, "Getting profile", err, http.StatusInternalServerError)
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
				Error(w, "Getting profiles", err, http.StatusInternalServerError)
				return
			}

			if err := templates.Profiles(profiles).Render(w); err != nil {
				Error(w, "Getting profiles", err, http.StatusInternalServerError)
				return
			}
		})

		r.Route("/{accountID}", func(r chi.Router) {
			r.Patch("/settings", func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()

				accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
				imageSize := r.FormValue("imageSize")
				if err := repo.UpdateAccountImageSize(r.Context(), accountID, imageSize); err != nil {
					Error(w, "Updating account image size", err, http.StatusInternalServerError)
					return
				}

				if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			r.Route("/profile", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

					account, err := repo.GetAccountByID(r.Context(), accountID)
					if err != nil {
						Error(w, "Getting account", err, http.StatusInternalServerError)
						return
					}

					profile, err := repo.GetProfileByAccountID(r.Context(), accountID)
					if err != nil {
						Error(w, "Getting profile", err, http.StatusInternalServerError)
						return
					}

					if err := templates.Profile(account, profile).Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

					r.ParseForm()

					displayName := r.FormValue("displayName")

					if _, err := repo.CreateProfile(r.Context(), accountID, displayName); err != nil {
						Error(w, "Creating profile", err, http.StatusInternalServerError)
						return
					}

					w.Header().Add("HX-Trigger", "profile-update")
					if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				})

				r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))

					r.ParseForm()

					displayName := r.FormValue("displayName")

					if err := repo.UpdateProfileDisplayName(r.Context(), accountID, displayName); err != nil {
						Error(w, "Updating profile", err, http.StatusInternalServerError)
						return
					}

					w.Header().Add("HX-Trigger", "profile-update")
					if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				})

				r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
					accountID := uuid.MustParse(chi.URLParam(r, "accountID"))
					if err := repo.DeleteProfile(r.Context(), accountID); err != nil {
						Error(w, "Deleting profile", err, http.StatusInternalServerError)
						return
					}

					w.Header().Add("HX-Trigger", "profile-update")
					if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
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
				Error(w, "Getting profile", err, http.StatusInternalServerError)
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
