package kroger

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
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

type GetAuthorizationRequest struct {
	Scope       string
	RedirectURI string
}

func (r GetAuthorizationRequest) applyToParams(params url.Values) {
	params.Add("scope", r.Scope)
	params.Add("redirect_uri", r.RedirectURI)
	params.Add("response_type", "code") // Hard coded
	// params.Add("state", randomState) // TODO: automate random value
}

type GetAuthorizationResponse struct {
	ExpiresIn    int64  `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

func (c *AuthorizationClient) GetAuthorization(ctx context.Context, request GetAuthorizationRequest) (*GetAuthorizationResponse, error) {
	var response GetAuthorizationResponse
	if err := c.client.Do(
		ctx,
		http.MethodGet,
		AuthorizationCodeEndpoint,
		request,
		&response,
		WithParam("client_id", c.clientID),
	); err != nil {
		return nil, err
	}
	return &response, nil
}

type PostTokenRequest struct {
	GrantType   string `json:"grant_type"`
	Scope       string `json:"scope"`
	Code        string `json:"code"`
	RedirectURI string `json:"redirest_uri"`
}

func (r PostTokenRequest) applyToParams(params url.Values) {
	params.Add("grant_type", r.GrantType)
	params.Add("scope", r.Scope)
}

type PostTokenResponse struct {
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`

	Error            string                 `json:"error"`
	ErrorDescription string                 `json:"error_description"`
	Errors           map[string]interface{} `json:"errors"`
}

func (c *AuthorizationClient) PostToken(ctx context.Context, request PostTokenRequest) (*PostTokenResponse, error) {
	var response PostTokenResponse
	if err := c.client.Do(
		ctx,
		http.MethodPost,
		AccessTokenEndpoint,
		request,
		&response,
		WithAuth(fmt.Sprintf("Basic %s", c.clientAuthorization)),
	); err != nil {
		return nil, err
	}
	return &response, nil
}
