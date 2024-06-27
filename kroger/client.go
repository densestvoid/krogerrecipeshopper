package kroger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type krogerClient struct {
	httpClient  *http.Client
	environment string
}

type apiRequest interface {
	applyToParams(params url.Values)
}

type option func(http.Header, url.Values)

func WithAuth(value string) option {
	return func(h http.Header, values url.Values) {
		h.Set("Authorization", value)
	}
}

func WithParam(key, value string) option {
	return func(h http.Header, values url.Values) {
		values.Add(key, value)
	}
}

func (c *krogerClient) Do(ctx context.Context, method string, endpoint string, request apiRequest, response any, options ...option) error {
	path, err := url.JoinPath(c.environment, endpoint)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return err
	}
	// httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	values := httpReq.URL.Query()
	if request != nil {
		request.applyToParams(values)
	}
	for _, opt := range options {
		opt(httpReq.Header, values)
	}
	httpReq.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(bs))

	return json.NewDecoder(resp.Body).Decode(&response)
}
