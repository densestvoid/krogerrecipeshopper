package kroger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	LocationsEndpoint = "/v1/locations"
)

type LocationsClient struct {
	client      *krogerClient
	accessToken string
}

func NewLocationsClient(client *http.Client, environment, accessToken string) *LocationsClient {
	return &LocationsClient{
		client: &krogerClient{
			httpClient:  client,
			environment: environment,
		},
		accessToken: accessToken,
	}
}

func (client *LocationsClient) auth() string {
	return fmt.Sprintf("Bearer %s", client.accessToken)
}

type GetLocationsRequest interface {
	getLocations(params url.Values)
}

type getLocationsRequest struct {
	GetLocationsRequest
}

func (r *getLocationsRequest) WriteHTTPRequest(req *http.Request) error {
	params := req.URL.Query()
	r.GetLocationsRequest.getLocations(params)
	req.URL.RawQuery = params.Encode()
	return nil
}

type GetLocationsByIDRequest struct {
	LocationIDs []string
}

func (r *GetLocationsByIDRequest) getLocations(params url.Values) {
	params.Add("filter.locationId", strings.Join(r.LocationIDs, ","))
}

type GetLocationsWithFiltersRequest struct {
	Chain          string
	Departments    []string // must be comma separated
	GeographicArea GeographicArea
}

func (r *GetLocationsWithFiltersRequest) getLocations(params url.Values) {
	params.Add("filter.chain", r.Chain)
	params.Add("filter.department", strings.Join(r.Departments, ","))

	if r.GeographicArea != nil {
		r.GeographicArea.geographicArea(params)
	}
}

type GeographicArea interface {
	geographicArea(params url.Values)
}

type GetLocationsWithZipCodeRequest struct {
	ZipCode       int
	RadiusInMiles int
}

func (r *GetLocationsWithZipCodeRequest) geographicArea(params url.Values) {
	params.Add("filter.zipCode.near", strconv.FormatInt(int64(r.ZipCode), 10))
	params.Add("filter.radiusInMiles", strconv.FormatFloat(float64(r.RadiusInMiles), 'f', -1, 32))
}

type GetLocationsWithLatitudeRequest struct {
	Latitude      float32
	RadiusInMiles int
}

func (r *GetLocationsWithLatitudeRequest) geographicArea(params url.Values) {
	params.Add("filter.lat.near", strconv.FormatFloat(float64(r.Latitude), 'f', -1, 32))
	params.Add("filter.radiusInMiles", strconv.FormatFloat(float64(r.RadiusInMiles), 'f', -1, 32))
}

type GetLocationsWithLongitudeRequest struct {
	Longitude     float32
	RadiusInMiles int
}

func (r *GetLocationsWithLongitudeRequest) geographicArea(params url.Values) {
	params.Add("filter.long.near", strconv.FormatFloat(float64(r.Longitude), 'f', -1, 32))
	params.Add("filter.radiusInMiles", strconv.FormatFloat(float64(r.RadiusInMiles), 'f', -1, 32))
}

type GetLocationsWithLatitudeAndLongitudeRequest struct {
	Latitude      float32
	Longitude     float32
	RadiusInMiles int
}

func (r *GetLocationsWithLatitudeAndLongitudeRequest) geographicArea(params url.Values) {
	lat := strconv.FormatFloat(float64(r.Latitude), 'f', -1, 32)
	long := strconv.FormatFloat(float64(r.Longitude), 'f', -1, 32)
	params.Add("filter.latLong.near", fmt.Sprintf("%s,%s", lat, long))
	params.Add("filter.radiusInMiles", strconv.FormatFloat(float64(r.RadiusInMiles), 'f', -1, 32))
}

type GetLocationsResponse struct {
	Meta      Meta       `json:"meta"`
	Locations []Location `json:"data"`
}

func (c *LocationsClient) GetLocations(ctx context.Context, request GetLocationsRequest) (*GetLocationsResponse, error) {
	var response GetLocationsResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		LocationsEndpoint,
		&getLocationsRequest{request},
		&HTTPResponseJSONParser{&response},
		WithAuth(c.auth()),
	); err != nil {
		return nil, err
	}
	return &response, nil
}

type GetLocationRequest struct {
	LocationID int
}

type GetLocationResponse struct {
	Meta     Meta     `json:"meta"`
	Location Location `json:"data"`
}

func (c *LocationsClient) GetLocation(ctx context.Context, request GetLocationRequest) (*GetProductResponse, error) {
	path, err := url.JoinPath(ProductsEndpoint, strconv.FormatInt(int64(request.LocationID), 10))
	if err != nil {
		return nil, err
	}

	var response GetProductResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		path,
		nil,
		&HTTPResponseJSONParser{&response},
		WithAuth(c.auth()),
	); err != nil {
		return nil, err
	}
	return &response, nil
}

type Location struct {
	LocationID     string `json:"locationId"`
	StoreNumber    string `json:"storeNumber"`
	DivisionNumber string `json:"divisionNumber"`
	Name           string
	Phone          string
	Chain          string
	Address        Address      `json:"address"`
	Departments    []Department `json:"departments"`
	Geolocation    Geolocation  `json:"geolocation"`
	Hours          WeeklyHours  `json:"hours"`
}

type Address struct {
	Line1   string `json:"addressLine1"`
	Line2   string `json:"addressLine2"`
	City    string
	County  string
	State   string
	ZipCode string
}

type Department struct {
	DepartmentID string `json:"departmentId"`
	Name         string
	Phone        string
	Hours        WeeklyHours `json:"hours"`
}

type WeeklyHours struct {
	Open24Hours bool       `json:"Open24"`
	GMTOffset   string     `json:"gmtOffset"`
	Timezone    string     `json:"timezone"`
	Monday      DailyHours `json:"monday"`
	Tuesday     DailyHours `json:"tuesday"`
	Wednesday   DailyHours `json:"wednesday"`
	Thursday    DailyHours `json:"thursday"`
	Friday      DailyHours `json:"friday"`
	Saturday    DailyHours `json:"saturday"`
	Sunday      DailyHours `json:"sunday"`
}

type DailyHours struct {
	Open        string `json:"open"`
	Close       string `json:"close"`
	Open24Hours bool   `json:"open24"`
}

type Geolocation struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}
