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
// Authentication uses per-session OIDC device flow: the first tool call
// returns an activation URL, the user authenticates in a browser, and
// subsequent calls use the cached token. Callers that already have a
// Bearer token (CI, service accounts) can pass it via the Authorization
// header to skip device flow entirely.
func serveHTTP(apiURL string) {
	listenAddr := os.Getenv("HIVETRACK_MCP_LISTEN")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	externalURL := os.Getenv("HIVETRACK_MCP_EXTERNAL_URL")
	if externalURL == "" {
		externalURL = "http://localhost" + listenAddr
	}
	externalURL = strings.TrimRight(externalURL, "/")

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s transport=http listen=%s external=%s\n", apiURL, listenAddr, externalURL)

	sessions := newSessionAuthManager(apiURL)
	client := htmcp.NewClient(apiURL, sessions)

	// Clean up session auth state when the MCP session ends.
	hooks := &server.Hooks{}
	hooks.AddOnUnregisterSession(func(_ context.Context, sess server.ClientSession) {
		sessions.removeSession(sess.SessionID())
	})

	s := htmcp.NewServer(client, server.WithHooks(hooks))

	mcpHandler := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
		// Extract optional Bearer token from the HTTP request into context.
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
				return context.WithValue(ctx, bearerTokenKey{}, token)
			}
			return ctx
		}),
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)

	proxy := newOAuthProxy(apiURL, externalURL)
	proxy.RegisterRoutes(mux)

	fmt.Fprintf(os.Stderr, "[mcp] serving HTTP on %s/mcp\n", listenAddr)
	if err := (&http.Server{Addr: listenAddr, Handler: mux}).ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
