package kroger

import (
	"bytes"
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

type HTTPRequestWriter interface {
	WriteHTTPRequest(*http.Request) error
}

type HTTPResponseParser interface {
	ParseHTTPResponse(*http.Response) error
}

func (c *krogerClient) Do(ctx context.Context, method string, endpoint string, request HTTPRequestWriter, response HTTPResponseParser, options ...option) error {
	path, err := url.JoinPath(c.environment, endpoint)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return err
	}

	if request != nil {
		if err := request.WriteHTTPRequest(httpReq); err != nil {
			return err
		}
	}

	values := httpReq.URL.Query()
	for _, opt := range options {
		opt(httpReq.Header, values)
	}
	httpReq.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if response != nil {
		return response.ParseHTTPResponse(resp)
	}
	return nil
}

type HTTPResponseBytesParser struct {
	Bytes []byte
}

func (p *HTTPResponseBytesParser) ParseHTTPResponse(resp *http.Response) error {
	// Get the body
	var err error
	p.Bytes, err = io.ReadAll(resp.Body)
	return err
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total int `json:"total"`
	Start int `json:"start"`
	Limit int `json:"limit"`
}

func (p Pagination) Prev() int {
	return max(p.Start-p.Limit, 0)
}

func (p Pagination) Next() int {
	return min(p.Start+p.Limit, p.Total)
}

type AuthError struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("Auth error: %s - %s", e.ErrorName, e.ErrorDescription)
}

type apiErrors struct {
	Errors APIError `json:"errors"`
}

type APIError struct {
	Timestamp int    `json:"timestamp"`
	Code      string `json:"code"`
	Reason    string `json:"reason"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s - %s", e.Timestamp, e.Code, e.Reason)
}

type HTTPResponseJSONParser struct {
	value any
}

func (p *HTTPResponseJSONParser) ParseHTTPResponse(resp *http.Response) error {
	// Get the body
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Resettable read buffer
	bytesReader := bytes.NewReader(bs)
	decoder := json.NewDecoder(bytesReader)
	decoder.DisallowUnknownFields()

	// Check for auth error
	var authError AuthError
	if err := decoder.Decode(&authError); err == nil {
		return &authError
	}

	// Reset
	if _, err := bytesReader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Check for API error
	var apiErrors apiErrors
	if err := decoder.Decode(&apiErrors); err == nil {
		return &apiErrors.Errors
	}

	// Reset
	if _, err := bytesReader.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Decode the response
	return decoder.Decode(&p.value)
}
