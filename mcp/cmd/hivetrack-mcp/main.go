package main

import (
	"context"
	"encoding/json"
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

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s transport=http listen=%s\n", apiURL, listenAddr)

	// Fetch OIDC metadata from the Hivetrack API so we can serve it at
	// /.well-known/oauth-authorization-server (MCP spec requirement).
	oauthMeta, err := fetchOAuthMetadata(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] warning: could not fetch OIDC metadata: %v\n", err)
		fmt.Fprintf(os.Stderr, "[mcp] OAuth discovery will be unavailable\n")
	}

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

	if oauthMeta != nil {
		metaJSON, _ := json.Marshal(oauthMeta)
		mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(metaJSON)
		})
	}

	fmt.Fprintf(os.Stderr, "[mcp] serving HTTP on %s/mcp\n", listenAddr)
	if err := (&http.Server{Addr: listenAddr, Handler: mux}).ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}

// fetchOAuthMetadata fetches the OIDC provider's discovery document via the
// Hivetrack API and returns an RFC 8414 OAuth Authorization Server Metadata
// object suitable for /.well-known/oauth-authorization-server.
func fetchOAuthMetadata(apiURL string) (map[string]any, error) {
	providerCfg, err := htclient.FetchOIDCProviderConfig(apiURL)
	if err != nil {
		return nil, err
	}

	oidcDoc, err := htclient.FetchOIDCDiscovery(providerCfg.Authority)
	if err != nil {
		return nil, err
	}

	meta := map[string]any{
		"issuer":                 oidcDoc["issuer"],
		"authorization_endpoint": oidcDoc["authorization_endpoint"],
		"token_endpoint":         oidcDoc["token_endpoint"],
		"registration_endpoint":  oidcDoc["registration_endpoint"],
	}
	copyIfPresent := func(key string, fallback any) {
		if v, ok := oidcDoc[key]; ok {
			meta[key] = v
		} else if fallback != nil {
			meta[key] = fallback
		}
	}
	copyIfPresent("response_types_supported", []string{"code"})
	copyIfPresent("code_challenge_methods_supported", []string{"S256"})
	copyIfPresent("token_endpoint_auth_methods_supported", nil)
	copyIfPresent("scopes_supported", nil)
	copyIfPresent("grant_types_supported", nil)

	return meta, nil
}
