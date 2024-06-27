package kroger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FulfillmentFilter string

const (
	ProductsEndpoint = "/v1/products"

	FulfillmentAvailableInStore FulfillmentFilter = "ais"
	FulfillmentCurbsidePickup   FulfillmentFilter = "csp"
	FulfillmenteliveryToHome    FulfillmentFilter = "dth"
	FulfillmentShipToHome       FulfillmentFilter = "sth"
)

type ProductsClient struct {
	client      *krogerClient
	environment string
	accessToken string
}

func NewProductsClient(client *http.Client, environment, accessToken string) *ProductsClient {
	return &ProductsClient{
		client: &krogerClient{
			httpClient:  client,
			environment: environment,
		},
		environment: environment,
		accessToken: accessToken,
	}
}

func (client *ProductsClient) auth() string {
	return fmt.Sprintf("Bearer %s", client.accessToken)
}

type GetProductsRequest struct {
	Filters    GetProductsFilters
	LocationID *string
}

func (r *GetProductsRequest) applyToParams(params url.Values) {
	if r.Filters != nil {
		r.Filters.getProductsFilters(params)
	}

	if r.LocationID != nil {
		params.Add("filter.locationId", *r.LocationID)
	}
}

type GetProductsFilters interface {
	getProductsFilters(parser url.Values)
}

type GetProductsByItemAndAvailabilityFilters struct {
	Term         string              // TODO: limit to 8 words, space separated. Remove punctuation?
	Brands       []string            // pipe separated
	Fulfillments []FulfillmentFilter // comma separated
	PageLimit    *int
	PageOffset   *int
}

func (r GetProductsByItemAndAvailabilityFilters) getProductsFilters(params url.Values) {
	params.Add("filter.term", r.Term)
	params.Add("filter.brand", strings.Join(r.Brands, "|"))

	var fulfillmentStrs []string
	for _, f := range r.Fulfillments {
		fulfillmentStrs = append(fulfillmentStrs, string(f))
	}
	params.Add("filter.fulfillment", strings.Join(fulfillmentStrs, ","))

	if r.PageLimit != nil {
		params.Add("filter.limit", strconv.FormatInt(int64(*r.PageLimit), 10))
	}

	if r.PageOffset != nil {
		params.Add("filter.start", strconv.FormatInt(int64(*r.PageOffset), 10))
	}
}

type GetProductsByIDsFilter struct {
	ProductIDs []int // comma separated
}

func (r *GetProductsByIDsFilter) getProductsFilters(params url.Values) {
	var productIDs []string
	for _, id := range r.ProductIDs {
		productIDs = append(productIDs, strconv.FormatInt(int64(id), 10))
	}
	params.Add("filter.productId", strings.Join(productIDs, ","))
}

type GetProductsResponse struct {
	Products []Product              `json:"data"`
	Errors   map[string]interface{} `json:"errors"`
	Meta     map[string]interface{} `json:"meta"`

	// Timestamp int
	// Code      string
	// Reason    string
}

func (c *ProductsClient) GetProducts(ctx context.Context, request GetProductsRequest) (*GetProductsResponse, error) {
	var response GetProductsResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		ProductsEndpoint,
		&request,
		&response,
		WithAuth(c.auth()),
	); err != nil {
		return nil, err
	}
	return &response, nil
}

type GetProductRequest struct {
	ProductID  int
	LocationID *int
}

func (r *GetProductRequest) applyToParams(params url.Values) {
	if r.LocationID != nil {
		params.Add("filter.locationId", strconv.FormatInt(int64(r.ProductID), 10))
	}
}

type GetProductResponse struct {
	Product Product                `json:"data"`
	Errors  map[string]interface{} `json:"errors"`
	Meta    map[string]interface{} `json:"meta"`
}

func (c *ProductsClient) GetProduct(ctx context.Context, request GetProductRequest) (*GetProductResponse, error) {
	path, err := url.JoinPath(ProductsEndpoint, strconv.FormatInt(int64(request.ProductID), 10))
	if err != nil {
		return nil, err
	}

	var response GetProductResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		path,
		&request,
		&response,
		WithAuth(c.auth()),
	); err != nil {
		return nil, err
	}
	return &response, nil
}

type Product struct {
	ProductID       int             `json:"productId,string"`
	AisleLocations  []AisleLocation `json:"aisleLocations"`
	Brand           string          `json:"brand"`
	Categories      []string        `json:"categories"`
	CountryOrigin   string          `json:"countryOrigin"`
	Description     string          `json:"description"`
	Items           []Item          `json:"items"`
	ItemInformation ItemInformation `json:"itemInformation"`
	Temperature     Temperature     `json:"temperature"`
	Images          []Image         `json:"images"`
	UPC             int             `json:"upc,string"`
}

type AisleLocation struct {
	BayNumber          int    `json:"bayNumber,string"`
	Description        string `json:"description"`
	Number             int    `json:"number,string"`
	NumberOfFacings    int    `json:"numberOfFacings,string"`
	SequenceNumber     int    `json:"sequenceNumber,string"`
	Side               string `json:"side"`
	ShelfNumber        int    `json:"shelfNumber,string"`
	ShelfPositionInBay int    `json:"shelfPositionInBay,string"`
}

type Item struct {
	ItemID        int         `json:"itemId,string"`
	Inventory     Inventory   `json:"inventory"`
	Favorite      bool        `json:"favorite"`
	Fulfillment   Fulfillment `json:"fulfillment"`
	Price         Price       `json:"price"`
	NationalPrice Price       `json:"nationalPrice"`
	Size          string      `json:"size"`
	SoldBy        string      `json:"soldBy"`
}

type Inventory struct {
	StockLevel string `json:"stockLevel"`
}

type Fulfillment struct {
	Curbside   bool `json:"curbside"`
	Delivery   bool `json:"delivery"`
	InStore    bool `json:"instore"`
	ShipToHome bool `json:"shiptohome"`
}

type Price struct {
	Regular                float32 `json:"regular"`
	Promo                  float32 `json:"promo"`
	RegularPerUnitEstimate float32 `json:"regularPerUnitEstimate"`
	PromoPerUnitEstimate   float32 `json:"promoPerUnitEstimate"`
}

type ItemInformation struct {
	Depth  float32 `json:"depth,string"`
	Height float32 `json:"height,string"`
	Width  float32 `json:"width,string"`
}

type Temperature struct {
	Indicator     string `json:"indicator"`
	HeatSensitive bool   `json:"heatSensitive"`
}

type Image struct {
	ID          string      `json:"id"`
	Perspective string      `json:"perspective"`
	Default     bool        `json:"default"`
	Sizes       []ImageSize `json:"sizes"`
}

type ImageSize struct {
	ID   string `json:"id"`
	Size string `json:"size"`
	URL  string `json:"url"`
}
