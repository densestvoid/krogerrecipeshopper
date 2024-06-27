package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/kroger"
)

const (
	ClientID     = "recipe-shopper-276be53a09a3bb09150aef03b5783ebc7840425982417759754"
	ClientSecret = "L6jkJmLdTBysOzNFABmKE06qa-5qwUW_tEXga1g-"
)

func main() {
	authClient := kroger.NewAuthorizationClient(http.DefaultClient, kroger.PublicEnvironment, ClientID, ClientSecret)

	resp, err := authClient.GetAuthorization(context.Background(), kroger.GetAuthorizationRequest{
		Scope:       kroger.ScopeCartBasicWrite,
		RedirectURI: "https://densestvoid.dev",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	// postTokenResp, err := authClient.PostToken(context.Background(), kroger.PostTokenRequest{
	// 	GrantType: kroger.GrantTypeClientCredentials,
	// 	Scope:     kroger.ScopeProductCompact,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// locationsClient := kroger.NewLocationsClient(http.DefaultClient, kroger.PublicEnvironment, postTokenResp.AccessToken)

	// getLocationsResp, err := locationsClient.GetLocations(context.Background(), &kroger.GetLocationsWithFiltersRequest{
	// 	GeographicArea: &kroger.GetLocationsWithZipCodeRequest{
	// 		ZipCode:       45248,
	// 		RadiusInMiles: 10,
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// productClient := kroger.NewProductsClient(http.DefaultClient, kroger.PublicEnvironment, postTokenResp.AccessToken)

	// getProductsResp, err := productClient.GetProducts(context.Background(), kroger.GetProductsRequest{
	// 	Filters: kroger.GetProductsByItemAndAvailabilityFilters{
	// 		Term: "milk",
	// 	},
	// 	LocationID: &getLocationsResp.Locations[0].LocationID,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// for _, product := range getProductsResp.Products {
	// 	for _, item := range product.Items {
	// 		fmt.Printf("%s: %f\n", product.Description, item.Price.Regular)
	// 	}
	// }
}
