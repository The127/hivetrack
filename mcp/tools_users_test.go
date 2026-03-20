package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetCurrentUser_whenCalled_callsUsersMeEndpoint(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/users/me" {
			t.Errorf("expected /api/v1/users/me, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"abc-123","email":"user@example.com","is_admin":false}`))
	}))
	defer srv.Close()

	handler := makeGetCurrentUser(NewClient(srv.URL, "tok"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected API to be called")
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
}

func TestGetCurrentUser_whenCalled_returnsEmailAndAdminStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"abc-123","email":"user@example.com","is_admin":true}`))
	}))
	defer srv.Close()

	handler := makeGetCurrentUser(NewClient(srv.URL, "tok"))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if text == "" {
		t.Fatal("expected non-empty result text")
	}
	// should contain email and admin indicator
	if !contains(text, "user@example.com") {
		t.Errorf("expected email in result, got: %s", text)
	}
	if !contains(text, "admin") {
		t.Errorf("expected admin status in result, got: %s", text)
	}
}
