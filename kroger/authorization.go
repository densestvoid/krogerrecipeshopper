package kroger

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	AuthorizationCodeEndpoint = "/v1/connect/oauth2/authorize"
	AccessTokenEndpoint       = "/v1/connect/oauth2/token"

	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeRefreshToken      = "refresh_token"

	ScopeProductCompact = "product.compact"
	ScopeCartBasicWrite = "cart.basic:write"
)

type AuthorizationClient struct {
	client              *krogerClient
	environment         string
	clientID            string
	clientAuthorization string
}

func NewAuthorizationClient(client *http.Client, environment, clientID, clientSecret string) *AuthorizationClient {
	return &AuthorizationClient{
		client: &krogerClient{
			httpClient:  client,
			environment: environment,
		},
		environment:         PublicEnvironment,
		clientID:            clientID,
		clientAuthorization: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))),
	}
}

type postTokenRequest struct {
	Credentials Credentials
}

func (r *postTokenRequest) WriteHTTPRequest(req *http.Request) error {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	values := url.Values{}
	r.Credentials.credentials(values)
	req.Body = io.NopCloser(strings.NewReader(values.Encode()))
	return nil
}

type Credentials interface {
	credentials(url.Values)
}

type AuthorizationCode struct {
	Code        string
	RedirectURI string
}

func (c AuthorizationCode) credentials(values url.Values) {
	values.Add("grant_type", GrantTypeAuthorizationCode)
	values.Add("code", c.Code)
	values.Add("redirect_uri", c.RedirectURI)
}

type ClientCredentials struct {
	Scope string
}

func (c ClientCredentials) credentials(values url.Values) {
	values.Add("grant_type", GrantTypeClientCredentials)
	values.Add("scope", c.Scope)
}

type RefreshToken struct {
	RefreshToken string
}

func (c RefreshToken) credentials(values url.Values) {
	values.Add("grant_type", GrantTypeRefreshToken)
	values.Add("refresh_token", c.RefreshToken)
}

type PostTokenResponse struct {
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

func (c *AuthorizationClient) PostToken(ctx context.Context, creds Credentials) (*PostTokenResponse, error) {
	var token PostTokenResponse
	if err := c.client.Do(
		ctx,
		http.MethodPost,
		AccessTokenEndpoint,
		&postTokenRequest{creds},
		&HTTPResponseJSONParser{&token},
		WithAuth(fmt.Sprintf("Basic %s", c.clientAuthorization)),
	); err != nil {
		return nil, err
	}
	return &token, nil
}
