package kroger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const (
	ProfileEndpoint = "/v1/identity/profile"
)

type IdentityClient struct {
	client      *krogerClient
	accessToken string
}

func NewIdentityClient(client *http.Client, environment, accessToken string) *IdentityClient {
	return &IdentityClient{
		client: &krogerClient{
			httpClient:  client,
			environment: environment,
		},
		accessToken: accessToken,
	}
}

func (client *IdentityClient) auth() string {
	return fmt.Sprintf("Bearer %s", client.accessToken)
}

type GetProfileResponse struct {
	Meta    Meta    `json:"meta"`
	Profile Profile `json:"data"`
}

func (c *IdentityClient) GetProfile(ctx context.Context) (*GetProfileResponse, error) {
	var token GetProfileResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		ProfileEndpoint,
		nil,
		&HTTPResponseJSONParser{&token},
		WithAuth(c.auth()),
	); err != nil {
		return nil, err
	}
	return &token, nil
}

type Profile struct {
	ID uuid.UUID `json:"id"`
}
