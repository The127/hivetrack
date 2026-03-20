package main

import (
	"context"
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

	token, err := htmcp.TryToken(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] auth error: %v\n", err)
		os.Exit(1)
	}

	var flow *htmcp.DeviceFlow
	if token == "" {
		flow, err = htmcp.InitDeviceFlow(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[mcp] failed to start device flow: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s\n", apiURL)

	client := htmcp.NewClient(apiURL, token)

	if flow != nil {
		authURL := flow.VerificationURIComplete
		if authURL == "" {
			authURL = flow.VerificationURI
		}
		client.SetAuthURL(authURL)
		go func() {
			tok, err := flow.WaitForToken(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "[mcp] device flow failed: %v\n", err)
				return
			}
			client.SetToken(tok)
			fmt.Fprintln(os.Stderr, "[mcp] authenticated")
		}()
	}

	s := htmcp.NewServer(client)

	fmt.Fprintln(os.Stderr, "[mcp] serving on stdio")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
