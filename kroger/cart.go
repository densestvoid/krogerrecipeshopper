package kroger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	CartAddEndpoint = "/v1/cart/add"
)

type CartClient struct {
	client      *krogerClient
	accessToken string
}

func NewCartClient(client *http.Client, environment, accessToken string) *CartClient {
	return &CartClient{
		client: &krogerClient{
			httpClient:  client,
			environment: environment,
		},
		accessToken: accessToken,
	}
}

func (client *CartClient) auth() string {
	return fmt.Sprintf("Bearer %s", client.accessToken)
}

const (
	ModalityPickup   = "PICKUP"
	ModalityDelivery = "DELIVERY"
)

type PutAddRequest struct {
	Items []PutAddProduct `json:"items"`
}

type PutAddProduct struct {
	ProductID string `json:"upc"`
	Quantity  int    `json:"quantity"`
	Modality  string `json:"modality"`
}

func (r *PutAddRequest) WriteHTTPRequest(req *http.Request) error {
	req.Header.Set("Content-Type", "application/json")
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	return nil
}

func (c *CartClient) PutAdd(ctx context.Context, req PutAddRequest) error {
	return c.client.Do(
		ctx,
		http.MethodPut,
		CartAddEndpoint,
		&req,
		nil,
		WithAuth(c.auth()),
	)
}
