package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/kroger"
)

type KrogerManager struct {
	productsClient  *kroger.ProductsClient
	locationsClient *kroger.LocationsClient
	cache           *data.Cache
}

func NewKrogerManager(productsClient *kroger.ProductsClient, locationsClient *kroger.LocationsClient, cache *data.Cache) *KrogerManager {
	return &KrogerManager{
		productsClient:  productsClient,
		locationsClient: locationsClient,
		cache:           cache,
	}
}

func (m *KrogerManager) GetProducts(ctx context.Context, productIDs ...string) (map[string]data.CacheProduct, error) {
	// Get products from cache
	cachedProdcuts, productIDMisses, err := m.cache.RetrieveKrogerProduct(ctx, productIDs...)
	if err != nil {
		return nil, err
	}

	var clientProducts []data.CacheProduct
	if len(productIDMisses) > 0 {
		// Get missing products from client
		productsResp, err := m.productsClient.GetProducts(ctx, kroger.GetProductsRequest{
			Filters: &kroger.GetProductsByIDsFilter{
				ProductIDs: productIDMisses,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, product := range productsResp.Products {
			clientProducts = append(clientProducts, KrogerProductToCacheProduct(product))
		}

		// Store products in cache
		if err := m.cache.StoreKrogerProduct(ctx, clientProducts...); err != nil {
			return nil, err
		}
	}

	// Return products
	var productsByID = map[string]data.CacheProduct{}
	for _, product := range slices.Concat(cachedProdcuts, clientProducts) {
		productsByID[product.ProductID] = product
	}
	return productsByID, nil
}

func (m *KrogerManager) GetLocation(ctx context.Context, locationID string) (data.CacheLocation, error) {
	// Get products from cache
	cachedLocation, err := m.cache.RetrieveKrogerLocation(ctx, locationID)
	if err != nil {
		return data.CacheLocation{}, err
	}

	if cachedLocation == nil {
		// Get missing location from client
		locationResp, err := m.locationsClient.GetLocation(ctx, kroger.GetLocationRequest{
			LocationID: locationID,
		})
		if err != nil {
			return data.CacheLocation{}, err
		}

		krogerLocation := locationResp.Location
		cachedLocation = &data.CacheLocation{
			LocationID: locationID,
			Name:       krogerLocation.Name,
			Address: fmt.Sprintf("%s %s %s %s %s",
				krogerLocation.Address.Line1,
				krogerLocation.Address.Line2,
				krogerLocation.Address.City,
				krogerLocation.Address.State,
				krogerLocation.Address.ZipCode,
			),
		}

		// Store location in cache
		if err := m.cache.StoreKrogerLocation(ctx, *cachedLocation); err != nil {
			return data.CacheLocation{}, err
		}
	}

	// Return location
	return *cachedLocation, nil
}

func KrogerProductToCacheProduct(product kroger.Product) data.CacheProduct {
	var size string
	for _, item := range product.Items {
		size = item.Size
		break
	}
	return data.CacheProduct{
		ProductID:   product.ProductID,
		Brand:       product.Brand,
		Description: product.Description,
		Size:        size,
		URL:         product.ProductPageURI,
	}
}
