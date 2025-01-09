package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewIngredientMux(repo *data.Repository) func(chi.Router) {
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
				token, err := r.Cookie("accessToken")
				if err != nil || token.Value == "" || token.Valid() != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, token.Value)
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
	}
}
