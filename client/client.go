// Package client provides a typed Go HTTP client for the Hivetrack API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// TokenFunc returns a Bearer token for authenticating API requests.
type TokenFunc func(ctx context.Context) (string, error)

// Client is a typed HTTP client for the Hivetrack API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	tokenFunc  TokenFunc
}

// New creates a new Hivetrack API client.
func New(baseURL string, tokenFunc TokenFunc) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenFunc: tokenFunc,
	}
}

// NewWithHTTPClient creates a new client with a custom http.Client (useful for testing).
func NewWithHTTPClient(baseURL string, tokenFunc TokenFunc, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		tokenFunc:  tokenFunc,
	}
}

// APIError represents an error returned by the Hivetrack API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hivetrack api: %d %s", e.StatusCode, e.Body)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.tokenFunc != nil {
		token, err := c.tokenFunc(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return json.RawMessage(`{}`), nil
	}

	return respBody, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, query, nil)
}

func (c *Client) post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) patch(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPatch, path, nil, body)
}

func (c *Client) delete(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// patchFields sends a PATCH with a struct containing Field[T] values,
// omitting absent fields and sending null for cleared fields.
func (c *Client) patchFields(ctx context.Context, path string, fields any) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodPatch, path, nil, fields)
}

// postFields sends a POST with a struct containing Field[T] values.
func (c *Client) postFields(ctx context.Context, path string, fields any) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodPost, path, nil, fields)
}

func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := marshalFields(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.tokenFunc != nil {
		token, err := c.tokenFunc(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return json.RawMessage(`{}`), nil
	}

	return respBody, nil
}
