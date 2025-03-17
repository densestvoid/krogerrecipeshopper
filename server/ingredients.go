package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/app"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func ProductImageLink(productID, imageSize string) string {
	return fmt.Sprintf("https://www.kroger.com/product/images/%s/front/%s", imageSize, productID)
}

func NewIngredientMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			listID := uuid.MustParse(chi.URLParam(r, "listID"))
			recipe, err := repo.GetRecipe(r.Context(), listID, authCookies.AccountID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.Ingredients(authCookies.AccountID, recipe).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			listID := uuid.MustParse(chi.URLParam(r, "listID"))

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			productID := r.FormValue("productID")

			staple := false
			if r.Form.Has("staple") {
				var err error
				if staple, err = strconv.ParseBool(r.FormValue("staple")); err != nil {
					http.Error(w, fmt.Sprintf("invalid staple value: %v", err), http.StatusBadRequest)
					return
				}
			}

			quantityPercent := 100
			if !staple {
				quantityFloat, err := strconv.ParseFloat(r.FormValue("quantity"), 64)
				if err != nil || quantityFloat <= 0 {
					http.Error(w, fmt.Sprintf("invalid quantity: %v", err), http.StatusBadRequest)
					return
				}
				quantityPercent = int(quantityFloat * 100)
			}

			if _, err := repo.GetIngredient(r.Context(), listID, productID); err != nil {
				// Doesn't exist, create it
				if err := repo.CreateIngredient(r.Context(), productID, listID, quantityPercent, staple); err != nil {
					http.Error(w, fmt.Sprintf("creating ingredient: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				// Exists, update it
				if err := repo.UpdateIngredient(r.Context(), productID, listID, quantityPercent, staple); err != nil {
					http.Error(w, fmt.Sprintf("updating ingredient: %v", err), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Add("HX-Trigger", "ingredient-update")
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/table", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			listID := uuid.MustParse(chi.URLParam(r, "listID"))
			recipe, err := repo.GetRecipe(r.Context(), listID, authCookies.AccountID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ingredients, err := repo.ListIngredients(r.Context(), listID)
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
				account, err := repo.GetAccountByID(r.Context(), authCookies.AccountID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
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
				productsByID, err := krogerManager.GetProducts(r.Context(), account.LocationID, productIDs...)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				for _, ingredient := range ingredients {
					product := productsByID[ingredient.ProductID]

					productURL, err := url.JoinPath(KrogerURL, product.URL)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					productURL, err = url.QueryUnescape(productURL)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					ingredientProducts = append(ingredientProducts, templates.Ingredient{
						Product: templates.Product{
							ProductID:   product.ProductID,
							Brand:       product.Brand,
							Description: product.Description,
							Size:        product.Size,
							ImageURL:    ProductImageLink(ingredient.ProductID, account.ImageSize),
							ProductURL:  productURL,
						},
						ListID:   ingredient.ListID,
						Quantity: ingredient.Quantity,
						Staple:   ingredient.Staple,
					})
				}
			}

			if err := templates.IngredientsTable(authCookies.AccountID, recipe.AccountID, ingredientProducts).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Route("/{productID}", func(r chi.Router) {
			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				listID := uuid.MustParse(chi.URLParam(r, "listID"))
				productID := chi.URLParam(r, "productID")
				var ingredient data.Ingredient
				if productID != "" {
					var err error
					ingredient, err = repo.GetIngredient(r.Context(), listID, productID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if err := templates.IngredientDetailsModalContent(listID, ingredient).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				listID := uuid.Must(uuid.Parse(chi.URLParam(r, "listID")))
				productID := chi.URLParam(r, "productID")
				if err := repo.DeleteIngredient(r.Context(), productID, listID); err != nil {
					http.Error(w, fmt.Sprintf("deleting ingredient: %v", err), http.StatusInternalServerError)
					return
				}
				w.Header().Add("HX-Trigger", "ingredient-update")
				w.WriteHeader(http.StatusOK)
			})
		})
	}
}
