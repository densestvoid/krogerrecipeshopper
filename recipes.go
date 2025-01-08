package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates/pages"
)

func NewRecipesMux(repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := pages.Recipes().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			userID, err := GetUserIDRequestCookie(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

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

			if r.PostForm.Has("id") {
				id, err := uuid.Parse(r.PostForm.Get("id"))
				if err != nil {
					http.Error(w, "Invalid ID", http.StatusBadRequest)
					return
				}

				if err := repo.UpdateRecipe(r.Context(), data.Recipe{
					ID:          id,
					UserID:      userID,
					Name:        name,
					Description: description,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				_, err := repo.CreateRecipe(r.Context(), userID, name, description)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Add("HX-Trigger", "recipe-update")
		})
		r.Get("/table", func(w http.ResponseWriter, r *http.Request) {
			userID, err := GetUserIDRequestCookie(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			recipes, err := repo.ListRecipes(r.Context(), data.ListRecipesFilterByUserID{UserID: userID})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := pages.RecipeTable(recipes).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Route("/{recipeID}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				var recipe *data.Recipe
				if recipeID, err := uuid.Parse(chi.URLParam(r, "recipeID")); err == nil {
					recipe, err = repo.GetRecipe(r.Context(), recipeID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if err := pages.RecipeDetailsForm(recipe).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))
				if err := repo.DeleteRecipe(r.Context(), recipeID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Add("HX-Trigger", "recipe-update")
			})
		})
	}
}