package server

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/densestvoid/krogerrecipeshopper/app"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

const KrogerURL = "https://www.kroger.com"

func NewProductsMux(config Config, repo *data.Repository, cache *data.Cache) func(chi.Router) {
	return func(r chi.Router) {
		r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
			authCookies, err := GetAuthCookies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

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

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			productsResp, err := productsClient.GetProducts(r.Context(), kroger.GetProductsRequest{
				Filters: kroger.GetProductsByItemAndAvailabilityFilters{
					Term: r.FormValue("search"),
				},
				LocationID: account.LocationID,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var cacheProducts []data.CacheProduct
			for _, product := range productsResp.Products {
				cacheProducts = append(cacheProducts, app.KrogerProductToCacheProduct(product))
			}

			// Store products in cache
			if err := cache.StoreKrogerProduct(r.Context(), cacheProducts...); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var products []templates.Product
			for _, product := range productsResp.Products {
				var size string
				for _, item := range product.Items {
					size = item.Size
					break
				}

				productURL, err := url.JoinPath(KrogerURL, product.ProductPageURI)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				productURL, err = url.QueryUnescape(productURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				products = append(products, templates.Product{
					ProductID:   product.ProductID,
					Brand:       product.Brand,
					Description: product.Description,
					Size:        size,
					ImageURL:    ProductImageLink(product.ProductID, account.ImageSize),
					ProductURL:  productURL,
				})
			}

			if err := templates.ProductsSearchTable(products).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
	}
}
