package recipes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewIngredientMux(config Config, repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			recipeID := uuid.MustParse(chi.URLParam(r, "recipeID"))
			recipe, err := repo.GetRecipe(r.Context(), recipeID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.Ingredients(recipe).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			recipeID := uuid.MustParse(chi.URLParam(r, "recipeID"))

			r.ParseForm()

			productID := r.FormValue("productID")
			quantityFloat, err := strconv.ParseFloat(r.FormValue("quantity"), 32)
			if err != nil || quantityFloat <= 0 {
				http.Error(w, "Invalid quantity", http.StatusBadRequest)
				return
			}
			quantityPercent := int(quantityFloat * 100)

			if _, err := repo.GetIngredient(r.Context(), recipeID, productID); err != nil {
				// Doesn't exist, create it
				if err := repo.CreateIngredient(r.Context(), productID, recipeID, quantityPercent); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				// Exists, update it
				if err := repo.UpdateIngredient(r.Context(), productID, recipeID, quantityPercent); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Add("HX-Trigger", "ingredient-update")
		})

		r.Get("/table", func(w http.ResponseWriter, r *http.Request) {
			recipeID := uuid.MustParse(chi.URLParam(r, "recipeID"))
			ingredients, err := repo.ListIngredients(r.Context(), recipeID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// hyrdate ingredients with product info
			productIDs := []string{}
			for _, ingredient := range ingredients {
				productIDs = append(productIDs, ingredient.ProductID)
			}

			ingredientProducts := []templates.Ingredient{}
			if len(productIDs) != 0 {
				authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
				authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{
					Scope: kroger.ScopeProductCompact,
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

				productsResp, err := productsClient.GetProducts(r.Context(), kroger.GetProductsRequest{
					Filters: &kroger.GetProductsByIDsFilter{
						ProductIDs: productIDs,
					},
					LocationID: nil,
				})

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				for _, ingredient := range ingredients {
					for _, product := range productsResp.Products {
						if ingredient.ProductID == product.ProductID {
							var imageURL string
						IMAGELOOP:
							for _, image := range product.Images {
								if image.Featured {
									for _, size := range image.Sizes {
										if size.Size == "thumbnail" {
											imageURL = size.URL
											break IMAGELOOP
										}
									}
								}
							}
							ingredientProducts = append(ingredientProducts, templates.Ingredient{
								Product: templates.Product{
									ProductID:   product.ProductID,
									Brand:       product.Brand,
									Description: product.Description,
									ImageURL:    imageURL,
								},
								RecipeID: ingredient.RecipeID,
								Quantity: ingredient.Quantity,
							})
							break
						}
					}
				}
			}

			if err := templates.IngredientsTable(ingredientProducts).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		r.Route("/{productID}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				recipeID := uuid.MustParse(chi.URLParam(r, "recipeID"))
				productID := chi.URLParam(r, "productID")
				var ingredient *data.Ingredient
				if productID != "" {
					var err error
					ingredient, err = repo.GetIngredient(r.Context(), recipeID, productID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if err := templates.IngredientDetailsForm(ingredient).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})

			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				recipeID := uuid.Must(uuid.Parse(chi.URLParam(r, "recipeID")))
				productID := chi.URLParam(r, "productID")
				if err := repo.DeleteIngredient(r.Context(), productID, recipeID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Add("HX-Trigger", "ingredient-update")
			})
		})
	}
}
