package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewRecipesMux(config Config, repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				Error(w, "Getting account id", err, http.StatusUnauthorized)
				return
			}

			if err := templates.Recipes(accountID).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Get("/explore", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.ExploreRecipes().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				Error(w, "Getting account id", err, http.StatusUnauthorized)
				return
			}

			r.ParseForm()

			name := r.FormValue("name")
			if name == "" {
				Error(w, "Name missing", nil, http.StatusBadRequest)
				return
			}

			description := r.FormValue("description")
			if description == "" {
				Error(w, "Description missing", nil, http.StatusBadRequest)
				return
			}

			if r.PostForm.Has("id") {
				id, err := uuid.Parse(r.PostForm.Get("id"))
				if err != nil {
					Error(w, "Parsing recipe id", err, http.StatusBadRequest)
					return
				}

				if err := repo.UpdateRecipe(r.Context(), data.Recipe{
					ID:          id,
					AccountID:   accountID,
					Name:        name,
					Description: description,
				}); err != nil {
					Error(w, "Updating recipe", err, http.StatusInternalServerError)
					return
				}
			} else {
				_, err := repo.CreateRecipe(r.Context(), accountID, name, description)
				if err != nil {
					Error(w, "Creating recipe", err, http.StatusInternalServerError)
					return
				}
			}

			w.Header().Add("HX-Trigger", "recipe-update")
			if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				Error(w, "Getting account id", err, http.StatusUnauthorized)
				return
			}

			r.ParseForm()
			filters := []data.ListRecipesFilter{}
			if r.Form.Has("accountID") {
				filters = append(filters, data.ListRecipesFilterByAccountID{AccountID: uuid.MustParse(r.Form.Get("accountID"))})
			}
			if r.Form.Has("name") && r.Form.Get("name") != "" {
				filters = append(filters, data.ListRecipesFilterByName{Name: r.FormValue("name")})
			}
			if len(filters) == 0 {
				if err := templates.RecipeTable(accountID, nil).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}

			recipes, err := repo.ListRecipes(r.Context(), filters...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.RecipeTable(accountID, recipes).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Route("/{recipeID}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
				if err != nil {
					Error(w, "Getting account id", err, http.StatusUnauthorized)
					return
				}

				var recipe *data.Recipe
				if recipeID, err := uuid.Parse(chi.URLParam(r, "recipeID")); err == nil {
					recipe, err = repo.GetRecipe(r.Context(), recipeID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if err := templates.RecipeDetailsForm(accountID, recipe).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
				if err != nil {
					Error(w, "Getting account id", err, http.StatusUnauthorized)
					return
				}
				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))

				if recipe, err := repo.GetRecipe(r.Context(), recipeID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				} else if recipe.AccountID != accountID {
					http.Error(w, "Can't delete recipes you didn't create", http.StatusBadRequest)
					return
				}

				if err := repo.DeleteRecipe(r.Context(), recipeID); err != nil {
					Error(w, "Deleting recipe", err, http.StatusInternalServerError)
					return
				}

				w.Header().Add("HX-Trigger", "recipe-update")
				if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})
			r.Route("/ingredients", NewIngredientMux(config, repo))
		})
	}
}
