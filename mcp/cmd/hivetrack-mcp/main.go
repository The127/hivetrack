package main

import (
	"fmt"
	"os"
	"time"

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

// serveHTTP runs the MCP server over HTTP (Streamable HTTP transport).
// Authenticates to the Hivetrack API using a static token (HIVETRACK_MCP_TOKEN).
func serveHTTP(apiURL string) {
	token := os.Getenv("HIVETRACK_MCP_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "[mcp] HIVETRACK_MCP_TOKEN is required in http transport mode")
		os.Exit(1)
	}

	listenAddr := os.Getenv("HIVETRACK_MCP_LISTEN")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s transport=http listen=%s\n", apiURL, listenAddr)

	provider := &htclient.StaticTokenProvider{
		Token: htclient.TokenCache{
			AccessToken: token,
			Expiry:      time.Now().Add(87600 * time.Hour),
			ServerURL:   apiURL,
		},
	}
	client := htmcp.NewClient(apiURL, provider)
	s := htmcp.NewServer(client)

	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
	)

	fmt.Fprintf(os.Stderr, "[mcp] serving HTTP on %s/mcp\n", listenAddr)
	if err := httpServer.Start(listenAddr); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
