package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

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

	if len(os.Args) > 1 && os.Args[1] == "login" {
		if err := htmcp.Login(apiURL); err != nil {
			fmt.Fprintf(os.Stderr, "[mcp] login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "[mcp] authenticated successfully")
		os.Exit(0)
	}

	tc, _ := htmcp.TryToken(apiURL) // best-effort; empty tokenCache if nothing valid

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s\n", apiURL)

	provider := htmcp.NewCachingTokenProvider(
		&htmcp.DeviceFlowProvider{BaseURL: apiURL},
		htmcp.RealClock,
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
