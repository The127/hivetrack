package main

import (
	"fmt"
	"os"

	htmcp "github.com/the127/hivetrack/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiURL := os.Getenv("HIVETRACK_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8086"
	}

	apiToken := os.Getenv("HIVETRACK_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "HIVETRACK_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[mcp] starting: url=%s token=%s...\n", apiURL, apiToken[:min(len(apiToken), 10)])

	client := htmcp.NewClient(apiURL, apiToken)
	s := htmcp.NewServer(client)

	fmt.Fprintln(os.Stderr, "[mcp] serving on stdio")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
