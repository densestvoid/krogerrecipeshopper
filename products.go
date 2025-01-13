package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewProductsMux(config Config) func(chi.Router) {
	return func(r chi.Router) {
		r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
			authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
			authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{
				Scope: kroger.ScopeProductCompact,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			productsClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)

			r.ParseForm()

			productsResp, err := productsClient.GetProducts(r.Context(), kroger.GetProductsRequest{
				Filters: kroger.GetProductsByItemAndAvailabilityFilters{
					Term: r.FormValue("search"),
				},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var products []templates.Product
			for _, product := range productsResp.Products {
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
				products = append(products, templates.Product{
					ProductID:   product.ProductID,
					Brand:       product.Brand,
					Description: product.Description,
					ImageURL:    imageURL,
				})
			}

			if err := templates.ProductsSearchTable(products).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
