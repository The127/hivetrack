package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListMilestones_whenSlugProvided_returnsMilestones(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"milestones": []map[string]any{{"id": "m1", "title": "v1.0", "issue_count": 5, "closed_issue_count": 3}},
		})
	}))
	defer srv.Close()

	handler := makeListMilestones(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := extractText(result)
	if !contains(text, "v1.0") || !contains(text, "3/5") {
		t.Errorf("expected milestone info, got: %s", text)
	}
}

func TestCreateMilestone_whenTitleProvided_createsMilestone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "ms-uuid"})
	}))
	defer srv.Close()

	handler := makeCreateMilestone(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "title": "v2.0"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := extractText(result)
	if !contains(text, "v2.0") || !contains(text, "ms-uuid") {
		t.Errorf("expected milestone info, got: %s", text)
	}
}

func TestUpdateMilestone_whenCloseProvided_closesMilestone(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeUpdateMilestone(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "milestone_id": "m1", "close": "true"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	text := extractText(result)
	if !contains(text, "closed") {
		t.Errorf("expected 'closed' in result, got: %s", text)
	}
}

func TestDeleteMilestone_whenIdProvided_deletesMilestone(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteMilestone(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "milestone_id": "m1"}

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
