package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	htclient "github.com/the127/hivetrack/client"
)

// Client is the MCP HTTP client. It wraps a typed client library instance
// and provides raw HTTP methods for backward compatibility during migration.
type Client struct {
	baseURL    string
	provider   htclient.TokenProvider
	httpClient *http.Client
	typed      *htclient.Client
}

// NewClient creates a new Hivetrack MCP client.
func NewClient(baseURL string, provider htclient.TokenProvider) *Client {
	base := strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:  base,
		provider: provider,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		typed: htclient.NewWithAuth(base, provider),
	}
}

// Typed returns the typed client library instance.
func (c *Client) Typed() *htclient.Client {
	return c.typed
}

// Raw HTTP methods — used by tests and the few remaining tools that build dynamic maps.

func (c *Client) get(path string, query url.Values) (json.RawMessage, error) {
	return c.do("GET", path, query, nil)
}

func (c *Client) post(path string, body any) (json.RawMessage, error) {
	return c.do("POST", path, nil, body)
}

func (c *Client) patch(path string, body any) (json.RawMessage, error) {
	return c.do("PATCH", path, nil, body)
}

func (c *Client) delete(path string) (json.RawMessage, error) {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) do(method, path string, query url.Values, body any) (json.RawMessage, error) {
	tc, err := c.provider.ProvideToken(context.Background())
	if err != nil {
		return nil, err
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tc.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	fmt.Fprintf(os.Stderr, "[mcp] %s %s\n", method, u)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] request failed: %v\n", err)
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[mcp] response: %d %s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return json.RawMessage(`{"ok":true}`), nil
	}

	return json.RawMessage(respBody), nil
}
