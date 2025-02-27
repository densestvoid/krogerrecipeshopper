package server

import (
	"net/http"
	"net/url"

	"github.com/densestvoid/krogerrecipeshopper/app"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
)

func NewShoppingListMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.ShoppingList().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/table", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			dataCartProducts, err := repo.ListCartProducts(r.Context(), authCookies.AccountID, &data.ListCartProductsNonStaples{})
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

				for _, dataCartProduct := range dataCartProducts {
					product := productsByID[dataCartProduct.ProductID]

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

					cartProducts = append(cartProducts, templates.CartProduct{
						ProductID:   product.ProductID,
						Brand:       product.Brand,
						Description: product.Description,
						Size:        product.Size,
						ImageURL:    ProductImageLink(dataCartProduct.ProductID, account.ImageSize),
						Quantity:    dataCartProduct.Quantity,
						Staple:      dataCartProduct.Staple,
						ProductURL:  productURL,
						Location:    product.Location,
					})
				}
			}

			if err := templates.ShoppingListTable(cartProducts).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
	}
}
