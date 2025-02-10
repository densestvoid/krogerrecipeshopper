package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/densestvoid/krogerrecipeshopper/kroger"
	"github.com/densestvoid/krogerrecipeshopper/templates"
	"github.com/go-chi/chi/v5"
)

const RadiusMiles = 5

func NewLocationsMux(config Config) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			zipCode, err := strconv.Atoi(r.Form.Get("zipCode"))
			if err != nil || zipCode <= 0 {
				http.Error(w, fmt.Sprintf("Invalid zip code %d: %v", zipCode, err), http.StatusBadRequest)
				return
			}

			authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, config.ClientID, config.ClientSecret)
			authResp, err := authClient.PostToken(r.Context(), kroger.ClientCredentials{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, authResp.AccessToken)
			locationsResp, err := locationsClient.GetLocations(r.Context(), &kroger.GetLocationsWithFiltersRequest{
				Chain:       "KROGER",
				Departments: nil,
				GeographicArea: &kroger.GetLocationsWithZipCodeRequest{
					ZipCode:       zipCode,
					RadiusInMiles: RadiusMiles,
				},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var locations []templates.Location
			for _, krogerLocation := range locationsResp.Locations {
				locations = append(locations, templates.Location{
					LocationID: krogerLocation.LocationID,
					Name:       krogerLocation.Name,
					Address: fmt.Sprintf("%s %s %s %s %s",
						krogerLocation.Address.Line1,
						krogerLocation.Address.Line2,
						krogerLocation.Address.City,
						krogerLocation.Address.State,
						krogerLocation.Address.ZipCode,
					),
				})
			}

			if err := templates.LocationsSearchTable(locations).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.LocationsSearch().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
