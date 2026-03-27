package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDeleteSprint_whenSlugAndSprintIdProvided_sendsDeleteRequest(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/sprints/sprint-uuid" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteSprint(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":      "my-proj",
		"sprint_id": "sprint-uuid",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected API to be called")
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
}

func TestDeleteSprint_whenSuccessful_confirmsSprintDeleted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteSprint(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":      "proj",
		"sprint_id": "sprint-uuid",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "deleted") {
		t.Errorf("expected deletion confirmation, got: %s", text)
	}
}

func TestUpdateSprint_whenForceProvided_sendsForceInBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeUpdateSprint(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":   "my-proj",
		"id":     "sprint-uuid",
		"status": "completed",
		"force":  true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", extractText(result))
	}
	if gotBody["force"] != true {
		t.Errorf("expected force=true in body, got: %v", gotBody["force"])
	}
	if gotBody["status"] != "completed" {
		t.Errorf("expected status=completed in body, got: %v", gotBody["status"])
	}
}

func TestUpdateSprint_whenMoveToSprintIdProvided_mapsToBackendField(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeUpdateSprint(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":               "proj",
		"id":                 "sprint-uuid",
		"status":             "completed",
		"move_to_sprint_id":  "next-sprint-uuid",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", extractText(result))
	}
	if gotBody["move_open_issues_to_sprint_id"] != "next-sprint-uuid" {
		t.Errorf("expected move_open_issues_to_sprint_id in body, got: %v", gotBody)
	}
}

func TestGetSprintBurndown_whenSlugAndSprintIdProvided_callsBurndownEndpoint(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/sprints/sprint-uuid/burndown" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":10,"start_remaining":10,"end_remaining":0,"points":[{"date":"2026-03-01","remaining":8},{"date":"2026-03-02","remaining":5}]}`))
	}))
	defer srv.Close()

	handler := makeGetSprintBurndown(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":      "my-proj",
		"sprint_id": "sprint-uuid",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected API to be called")
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
}

func TestGetSprintBurndown_whenDataReturned_formatsPointsWithDatesAndRemaining(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":10,"start_remaining":10,"end_remaining":2,"points":[{"date":"2026-03-01","remaining":8},{"date":"2026-03-02","remaining":5},{"date":"2026-03-03","remaining":2}]}`))
	}))
	defer srv.Close()

	handler := makeGetSprintBurndown(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":      "proj",
		"sprint_id": "sprint-uuid",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "2026-03-01") {
		t.Errorf("expected date in result, got: %s", text)
	}
	if !contains(text, "8") {
		t.Errorf("expected remaining count in result, got: %s", text)
	}
}
