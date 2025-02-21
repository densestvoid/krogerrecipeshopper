package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewRecipesMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if err := templates.Recipes(authCookies.AccountID).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/favorites", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.FavoriteRecipes().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/explore", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.ExploreRecipes().Render(w); err != nil {
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

			instructionType := r.FormValue("instruction-type")
			if instructionType == "" {
				http.Error(w, fmt.Sprintf("instruction type missing: %v", err), http.StatusBadRequest)
				return
			}

			instructions := r.FormValue("instructions")
			if instructionType != data.InstructionTypeNone && instructions == "" {
				http.Error(w, fmt.Sprintf("instructions missing: %v", err), http.StatusBadRequest)
				return
			}

			visibility := r.FormValue("visibility")
			if visibility == "" {
				http.Error(w, fmt.Sprintf("visibility missing: %v", err), http.StatusBadRequest)
				return
			}

			if r.PostForm.Has("id") {
				id, err := uuid.Parse(r.PostForm.Get("id"))
				if err != nil {
					http.Error(w, fmt.Sprintf("parsing recipe id: %v", err), http.StatusBadRequest)
					return
				}

				if err := repo.UpdateRecipe(r.Context(), data.Recipe{
					ID:              id,
					AccountID:       authCookies.AccountID,
					Name:            name,
					Description:     description,
					InstructionType: instructionType,
					Instructions:    instructions,
					Visibility:      visibility,
				}); err != nil {
					http.Error(w, fmt.Sprintf("updating recipe: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				_, err := repo.CreateRecipe(r.Context(), authCookies.AccountID, name, description, instructionType, instructions, visibility)
				if err != nil {
					http.Error(w, fmt.Sprintf("creating recipe: %v", err), http.StatusInternalServerError)
					return
				}
			}
			w.Header().Add("HX-Trigger", "recipe-update")
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

			filters := []data.ListRecipesFilter{}
			if r.Form.Has("accountID") {
				filters = append(filters, data.ListRecipesFilterByAccountID{AccountID: uuid.MustParse(r.Form.Get("accountID"))})
			}
			if r.Form.Has("name") && r.Form.Get("name") != "" {
				filters = append(filters, data.ListRecipesFilterByName{Name: r.FormValue("name")})
			}
			if r.Form.Has("favorites") {
				filters = append(filters, data.ListRecipesFilterByFavorites{})
			}
			slog.Info("", "form", r.Form)
			filters = append(filters, data.ListRecipesFilterByVisibilities{Visibilities: r.Form["visibility"]})
			if len(filters) == 0 {
				if err := templates.RecipeTable(authCookies.AccountID, nil).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}

			recipes, err := repo.ListRecipes(r.Context(), authCookies.AccountID, filters, []data.ListRecipesOrderBy{
				{Field: "name", Direction: "asc"},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.RecipeTable(authCookies.AccountID, recipes).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Route("/{recipeID}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				var recipe *data.Recipe
				if recipeID, err := uuid.Parse(chi.URLParam(r, "recipeID")); err == nil {
					recipe, err = repo.GetRecipe(r.Context(), recipeID, authCookies.AccountID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if err := templates.RecipeDetailsForm(authCookies.AccountID, recipe).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))

				if recipe, err := repo.GetRecipe(r.Context(), recipeID, authCookies.AccountID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				} else if recipe.AccountID != authCookies.AccountID {
					http.Error(w, "Can't delete recipes you didn't create", http.StatusBadRequest)
					return
				}

				if err := repo.DeleteRecipe(r.Context(), recipeID); err != nil {
					http.Error(w, fmt.Sprintf("deleting recipe: %v", err), http.StatusInternalServerError)
					return
				}

				w.Header().Add("HX-Trigger", "recipe-update")
				w.WriteHeader(http.StatusOK)
			})

			r.Post("/favorite", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))

				if err := repo.FavoriteRecipe(r.Context(), recipeID, authCookies.AccountID); err != nil {
					http.Error(w, fmt.Sprintf("adding favorite recipe: %v", err), http.StatusInternalServerError)
					return
				}

				if err := templates.FavoriteButton(recipeID, true).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Delete("/favorite", func(w http.ResponseWriter, r *http.Request) {
				authCookies, err := GetAuthCookies(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))

				if err := repo.UnfavoriteRecipe(r.Context(), recipeID, authCookies.AccountID); err != nil {
					http.Error(w, fmt.Sprintf("removing favorite recipe: %v", err), http.StatusInternalServerError)
					return
				}

				if err := templates.FavoriteButton(recipeID, false).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Route("/ingredients", NewIngredientMux(config, repo, cache))
		})
	}
}
