package server

import (
	"fmt"
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewListsMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if err := templates.Lists(authCookies.AccountID).Render(w); err != nil {
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

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			name := r.FormValue("name")
			if name == "" {
				http.Error(w, fmt.Sprintf("name missing: %v", err), http.StatusBadRequest)
				return
			}

			description := r.FormValue("description")

			if r.PostForm.Has("id") {
				id, err := uuid.Parse(r.PostForm.Get("id"))
				if err != nil {
					http.Error(w, fmt.Sprintf("parsing list id: %v", err), http.StatusBadRequest)
					return
				}

				if err := repo.UpdateList(r.Context(), data.List{
					ID:          id,
					AccountID:   authCookies.AccountID,
					Name:        name,
					Description: description,
				}); err != nil {
					http.Error(w, fmt.Sprintf("updating list: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				_, err := repo.CreateList(r.Context(), authCookies.AccountID, name, description)
				if err != nil {
					http.Error(w, fmt.Sprintf("creating list: %v", err), http.StatusInternalServerError)
					return
				}
			}
			w.Header().Add("HX-Trigger", "list-update")
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			filters := []data.ListListsFilter{}
			if r.Form.Has("name") && r.Form.Get("name") != "" {
				filters = append(filters, data.ListListsFilterByName{Name: r.FormValue("name")})
			}

			lists, err := repo.ListLists(r.Context(), authCookies.AccountID, filters, []data.ListListsOrderBy{
				{Field: "name", Direction: "asc"},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.ListTable(authCookies.AccountID, lists).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				var list data.List
				if listID, err := uuid.Parse(chi.URLParam(r, "id")); err == nil {
					list, err = repo.GetList(r.Context(), listID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if list.AccountID != uuid.Nil && list.AccountID != authCookies.AccountID {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				if err := templates.ListDetailsModalContent(authCookies.AccountID, list, false).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Get("/copy", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				var list data.List
				if listID, err := uuid.Parse(chi.URLParam(r, "id")); err == nil {
					list, err = repo.GetList(r.Context(), listID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if list.AccountID != authCookies.AccountID {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				if err := templates.ListDetailsModalContent(authCookies.AccountID, list, true).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Post("/copy", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				if err := r.ParseForm(); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				name := r.FormValue("name")
				if name == "" {
					http.Error(w, fmt.Sprintf("name missing: %v", err), http.StatusBadRequest)
					return
				}

				description := r.FormValue("description")

				listToCopyID, err := uuid.Parse(r.PostForm.Get("id"))
				if err != nil {
					http.Error(w, fmt.Sprintf("parsing list id: %v", err), http.StatusBadRequest)
					return
				}

				newListID, err := repo.CreateList(r.Context(), authCookies.AccountID, name, description)
				if err != nil {
					http.Error(w, fmt.Sprintf("creating new recipe: %v", err), http.StatusInternalServerError)
					return
				}

				ingredients, err := repo.ListIngredients(r.Context(), listToCopyID)
				if err != nil {
					http.Error(w, fmt.Sprintf("listing ingredients to copy: %v", err), http.StatusInternalServerError)
					return
				}

				for _, ingredient := range ingredients {
					if err := repo.CreateIngredient(r.Context(), ingredient.ProductID, newListID, ingredient.Quantity, ingredient.Staple); err != nil {
						http.Error(w, fmt.Sprintf("adding ingredient to copied list: %v", err), http.StatusInternalServerError)
						return
					}
				}

				w.Header().Add("HX-Trigger", "list-update")
				w.WriteHeader(http.StatusOK)
			})

			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				listID := uuid.Must(uuid.Parse(chi.URLParam(r, "id")))

				list, err := repo.GetList(r.Context(), listID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				} else if list.AccountID != authCookies.AccountID {
					http.Error(w, "Can't delete lists you didn't create", http.StatusBadRequest)
					return
				}

				if err := repo.DeleteList(r.Context(), list.ID); err != nil {
					http.Error(w, fmt.Sprintf("deleting list: %v", err), http.StatusInternalServerError)
					return
				}

				w.Header().Add("HX-Trigger", "list-update")
				w.WriteHeader(http.StatusOK)
			})

			r.Route("/ingredients", NewIngredientMux(config, repo, cache))
		})
	}
}
