package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListLabels_whenSlugProvided_callsLabelsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/labels" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"labels": []map[string]any{{"id": "l1", "name": "bug", "color": "#ff0000"}},
		})
	}))
	defer srv.Close()

	handler := makeListLabels(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
	text := extractText(result)
	if !contains(text, "bug") {
		t.Errorf("expected 'bug' in result, got: %s", text)
	}
}

func TestCreateLabel_whenNameAndColorProvided_createsLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"ID": "new-label-id"})
	}))
	defer srv.Close()

	handler := makeCreateLabel(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "name": "feature", "color": "#00ff00"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := extractText(result)
	if !contains(text, "feature") || !contains(text, "new-label-id") {
		t.Errorf("expected label info in result, got: %s", text)
	}
}

func TestDeleteLabel_whenLabelIdProvided_deletesLabel(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteLabel(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "label_id": "l1"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestUpdateLabel_whenFieldsProvided_updatesLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeUpdateLabel(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "label_id": "l1", "name": "renamed"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := extractText(result)
	if !contains(text, "renamed") {
		t.Errorf("expected 'renamed' in result, got: %s", text)
	}
}
