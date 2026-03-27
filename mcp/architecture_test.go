// Package mcp architecture tests enforce the boundary between the MCP server
// and the client library.
//
// Rules:
//   - MCP must NOT contain auth logic (device flow, token refresh, OIDC discovery)
//   - MCP must import client library for all API operations and auth
//   - Client library must NOT import MCP (no circular dependency)
//   - All tool handlers must use client.Typed() — no raw HTTP calls to Hivetrack API
package mcp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAuthLogicInMCP ensures that auth-related code lives in the client library,
// not in the MCP server. The MCP module should only configure and use auth providers.
func TestNoAuthLogicInMCP(t *testing.T) {
	// Auth patterns that must NOT appear in non-test MCP Go files.
	forbidden := []string{
		"DeviceAuthorizationEndpoint",
		"oidcDiscovery",
		"oidcProviderConfig",
		"deviceAuthResponse",
		"tokenResponse",
		"tryRefresh",
		"fetchOIDCEndpoints",
		"postFormJSON",
		"func saveCache",
		"func loadCache",
		"func cachePath",
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Errorf("mcp/%s contains auth logic %q — auth belongs in the client library", f, pattern)
			}
		}
	}
}

// TestMCPImportsClientLibrary ensures the MCP module imports the client library.
func TestMCPImportsClientLibrary(t *testing.T) {
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	clientImported := false
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		node, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", f, err)
		}
		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == "github.com/the127/hivetrack/client" {
				clientImported = true
			}
		}
	}

	if !clientImported {
		t.Error("MCP module must import github.com/the127/hivetrack/client")
	}
}

// TestNoRawHTTPCallsInToolHandlers ensures all tool handler files use the typed
// client (client.Typed()) rather than raw HTTP methods (client.get/post/patch/delete).
func TestNoRawHTTPCallsInToolHandlers(t *testing.T) {
	rawPatterns := []string{
		"client.get(",
		"client.post(",
		"client.patch(",
		"client.delete(",
	}

	files, err := filepath.Glob("tools_*.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		for _, pattern := range rawPatterns {
			if strings.Contains(content, pattern) {
				t.Errorf("mcp/%s contains raw HTTP call %q — use client.Typed() methods instead", f, pattern)
			}
		}
	}
}

// TestNoTokenCacheTypeInMCP ensures the MCP module doesn't define its own token types.
func TestNoTokenCacheTypeInMCP(t *testing.T) {
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	forbiddenTypes := map[string]bool{
		"tokenCache":    true,
		"TokenProvider": true,
		"Clock":         true,
	}

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		node, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", f, err)
		}
		for _, decl := range node.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if forbiddenTypes[ts.Name.Name] {
					t.Errorf("mcp/%s defines type %q — this belongs in the client library", f, ts.Name.Name)
				}
			}
		}
	}
}
