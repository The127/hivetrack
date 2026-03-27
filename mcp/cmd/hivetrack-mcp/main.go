package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	htclient "github.com/the127/hivetrack/client"
	htmcp "github.com/the127/hivetrack/mcp"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Fprintln(os.Stdout, "hivetrack-mcp "+version)
		os.Exit(0)
	}

	apiURL := os.Getenv("HIVETRACK_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8086"
	}

	transport := os.Getenv("HIVETRACK_MCP_TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}

	if len(os.Args) > 1 && os.Args[1] == "login" {
		if err := htclient.Login(apiURL); err != nil {
			fmt.Fprintf(os.Stderr, "[mcp] login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "[mcp] authenticated successfully")
		os.Exit(0)
	}

	switch transport {
	case "http":
		serveHTTP(apiURL)
	default:
		serveStdio(apiURL)
	}
}

// serveStdio runs the MCP server over stdin/stdout with OIDC device flow auth.
func serveStdio(apiURL string) {
	tc, _ := htclient.LoadTokenFile()

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s transport=stdio\n", apiURL)

	provider := htclient.NewCachingTokenProvider(
		&htclient.DeviceFlowProvider{BaseURL: apiURL},
		htclient.RealClock,
		apiURL,
		tc,
		0.1,
	)
	client := htmcp.NewClient(apiURL, provider)
	s := htmcp.NewServer(client)

	fmt.Fprintln(os.Stderr, "[mcp] serving on stdio")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}

// bearerTokenKey is the context key for the caller's Bearer token.
type bearerTokenKey struct{}

// serveHTTP runs the MCP server over HTTP (Streamable HTTP transport).
// The caller's OIDC Bearer token is extracted from the incoming HTTP request
// and passed through to the Hivetrack API. The MCP server itself has no
// credentials — it's a transparent proxy for authentication.
func serveHTTP(apiURL string) {
	listenAddr := os.Getenv("HIVETRACK_MCP_LISTEN")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s transport=http listen=%s\n", apiURL, listenAddr)

	// The token function reads the Bearer token from the request context.
	// Each MCP request carries the caller's own OIDC token through to the API.
	client := htmcp.NewClient(apiURL, htclient.TokenProviderFunc(func(ctx context.Context) (htclient.TokenCache, error) {
		token, ok := ctx.Value(bearerTokenKey{}).(string)
		if !ok || token == "" {
			return htclient.TokenCache{}, fmt.Errorf("no bearer token in request — pass Authorization header")
		}
		return htclient.TokenCache{AccessToken: token}, nil
	}))

	s := htmcp.NewServer(client)

	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
		// Extract the caller's Bearer token from the HTTP request into context.
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			auth := r.Header.Get("Authorization")
			token := strings.TrimPrefix(auth, "Bearer ")
			return context.WithValue(ctx, bearerTokenKey{}, token)
		}),
	)

	fmt.Fprintf(os.Stderr, "[mcp] serving HTTP on %s/mcp\n", listenAddr)
	if err := httpServer.Start(listenAddr); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
