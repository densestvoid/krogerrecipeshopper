package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/densestvoid/krogerrecipeshopper/app"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewT[T any](t T) *T {
	return &t
}

const KrogerCartURL = "https://www.kroger.com/shopping/cart"

func NewCartMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.Cart().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/table", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			dataCartProducts, err := repo.ListCartProducts(r.Context(), accountID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// hyrdate ingredients with product info
			productIDs := []string{}
			for _, cartProduct := range dataCartProducts {
				productIDs = append(productIDs, cartProduct.ProductID)
			}

			cartProducts := []templates.CartProduct{}
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
				locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

				krogerManager := app.NewKrogerManager(productsClient, locationsClient, cache)
				productsByID, err := krogerManager.GetProducts(r.Context(), productIDs...)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				account, err := repo.GetAccountByID(r.Context(), accountID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				for _, dataCartProduct := range dataCartProducts {
					product := productsByID[dataCartProduct.ProductID]
					cartProducts = append(cartProducts, templates.CartProduct{
						ProductID:   product.ProductID,
						Brand:       product.Brand,
						Description: product.Description,
						Size:        product.Size,
						ImageURL:    ProductImageLink(dataCartProduct.ProductID, account.ImageSize),
						Quantity:    dataCartProduct.Quantity,
						Staple:      dataCartProduct.Staple,
					})
				}
			}

			if err := templates.CartTable(cartProducts).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Post("/recipe/{recipeID}", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
				return
			}

			recipeID := uuid.MustParse(chi.URLParam(r, "recipeID"))
			ingredients, err := repo.ListIngredients(r.Context(), recipeID)
			if err != nil {
				http.Error(w, fmt.Sprintf("listing ingredients: %v", err), http.StatusInternalServerError)
				return
			}
			for _, ingredient := range ingredients {
				if err := repo.AddCartProduct(r.Context(), accountID, ingredient.ProductID, ingredient.Quantity, ingredient.Staple); err != nil {
					http.Error(w, fmt.Sprintf("adding cart product: %v", err), http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
		})

		// Set product quantity in users cart
		r.Put("/product", func(w http.ResponseWriter, r *http.Request) {
			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
				return
			}

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			productID := r.FormValue("productID")
			quantityFloat, err := strconv.ParseFloat(r.FormValue("quantity"), 64)
			if err != nil || quantityFloat <= 0 {
				http.Error(w, fmt.Sprintf("invalid quantity: %v", err), http.StatusBadRequest)
				return
			}
			quantityPercent := int(quantityFloat * 100)

			if err := repo.SetCartProduct(r.Context(), accountID, productID, &quantityPercent, nil); err != nil {
				http.Error(w, fmt.Sprintf("updating cart product: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Add("HX-Trigger", "cart-update")
			w.WriteHeader(http.StatusOK)
		})

		r.Route("/{productID}", func(r chi.Router) {
			// Get cart product details
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				productID := chi.URLParam(r, "productID")

				cartProduct, err := repo.GetCartProduct(r.Context(), accountID, productID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := templates.CartProductDetailsForm(cartProduct).Render(w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			// Remove product from users cart
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
				if err != nil {
					http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
					return
				}
				productID := chi.URLParam(r, "productID")

				if err := repo.RemoveCartProduct(r.Context(), accountID, productID); err != nil {
					http.Error(w, fmt.Sprintf("removing cart product: %v", err), http.StatusInternalServerError)
					return
				}

				w.Header().Add("HX-Trigger", "cart-update")
				w.WriteHeader(http.StatusOK)
			})

			// Update product to be included in the products sent to the kroger cart
			r.Post("/include", func(w http.ResponseWriter, r *http.Request) {
				accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
				if err != nil {
					http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
					return
				}
				productID := chi.URLParam(r, "productID")

				if err := repo.SetCartProduct(r.Context(), accountID, productID, nil, NewT(false)); err != nil {
					http.Error(w, fmt.Sprintf("updating cart product: %v", err), http.StatusInternalServerError)
					return
				}

				w.Header().Add("HX-Trigger", "cart-update")
				w.WriteHeader(http.StatusOK)
			})
		})

		r.Post("/checkout", func(w http.ResponseWriter, r *http.Request) {
			accessTokenCookie, err := r.Cookie("accessToken")
			if err != nil {
				http.Error(w, fmt.Sprintf("getting access token: %v", err), http.StatusUnauthorized)
				return
			}
			cartClient := kroger.NewCartClient(http.DefaultClient, kroger.PublicEnvironment, accessTokenCookie.Value)

			accountID, err := GetAccountIDFromRequestSessionCookie(repo, r)
			if err != nil {
				http.Error(w, fmt.Sprintf("getting account id: %v", err), http.StatusUnauthorized)
				return
			}

			cartProducts, err := repo.ListCartProducts(r.Context(), accountID, &data.ListCartProductsNonStaples{})
			if err != nil {
				http.Error(w, fmt.Sprintf("listing cart products: %v", err), http.StatusInternalServerError)
				return
			}

			var addProducts []kroger.PutAddProduct
			for _, cartProduct := range cartProducts {
				addProducts = append(addProducts, kroger.PutAddProduct{
					ProductID: cartProduct.ProductID,
					Quantity:  int(math.Ceil(float64(cartProduct.Quantity) / 100)),
					Modality:  kroger.ModalityPickup,
				})
			}

			if err := cartClient.PutAdd(r.Context(), kroger.PutAddRequest{
				Items: addProducts,
			}); err != nil {
				http.Error(w, fmt.Sprintf("adding products to kroger cart: %v", err), http.StatusInternalServerError)
				return
			}

			if err := repo.ClearCartProducts(r.Context(), accountID); err != nil {
				http.Error(w, fmt.Sprintf("removing cart products: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Add("HX-Trigger", "cart-update")
			w.WriteHeader(http.StatusOK)
		})
	}
}
