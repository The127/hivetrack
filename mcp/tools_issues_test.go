package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDeleteIssue_whenSlugAndNumberProvided_sendsDeleteRequest(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/issues/42" {
			t.Errorf("expected /api/v1/projects/my-proj/issues/42, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "my-proj", "number": float64(42)}

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

func TestDeleteIssue_whenSuccessful_confirmsDeletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "number": float64(7)}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "deleted") {
		t.Errorf("expected deletion confirmation, got: %s", text)
	}
}

func TestDeleteIssue_whenSlugMissing_returnsErrorWithoutCallingAPI(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeDeleteIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"number": float64(42)}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected API NOT to be called when slug missing")
	}
	if !result.IsError {
		t.Error("expected error result when slug missing")
	}
}

func TestRefineIssue_whenSlugAndNumberProvided_sendsPostToRefineEndpoint(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/issues/5/refine" {
			t.Errorf("expected /api/v1/projects/my-proj/issues/5/refine, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeRefineIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "my-proj", "number": float64(5)}

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

func TestRefineIssue_whenSuccessful_confirmsIssueMarkedRefined(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeRefineIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"slug": "proj", "number": float64(3)}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "refined") {
		t.Errorf("expected refinement confirmation, got: %s", text)
	}
}

func TestAddIssueLink_whenLinkTypeAndTargetProvided_sendsLinkBodyToLinksEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/issues/10/links" {
			t.Errorf("expected /api/v1/projects/my-proj/issues/10/links, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeAddIssueLink(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":          "my-proj",
		"number":        float64(10),
		"link_type":     "blocks",
		"target_number": float64(20),
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
	if gotBody["link_type"] != "blocks" {
		t.Errorf("expected link_type=blocks in body, got: %v", gotBody["link_type"])
	}
	if gotBody["target_number"] != float64(20) {
		t.Errorf("expected target_number=20 in body, got: %v", gotBody["target_number"])
	}
}

func TestAddIssueLink_whenSuccessful_confirmsLinkCreated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	handler := makeAddIssueLink(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":          "proj",
		"number":        float64(1),
		"link_type":     "relates_to",
		"target_number": float64(2),
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "link") {
		t.Errorf("expected link confirmation, got: %s", text)
	}
}

func TestSplitIssue_whenTitlesProvided_sendsArrayToSplitEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/my-proj/issues/3/split" {
			t.Errorf("expected /api/v1/projects/my-proj/issues/3/split, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"new_issues":[{"id":"x1","number":5},{"id":"x2","number":6}]}`))
	}))
	defer srv.Close()

	handler := makeSplitIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":   "my-proj",
		"number": float64(3),
		"titles": "First part,Second part",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
	titlesRaw, ok := gotBody["titles"]
	if !ok {
		t.Fatal("expected titles in body")
	}
	titles, ok := titlesRaw.([]interface{})
	if !ok {
		t.Fatalf("expected titles to be array, got %T", titlesRaw)
	}
	if len(titles) != 2 {
		t.Errorf("expected 2 titles, got %d", len(titles))
	}
}

func TestSplitIssue_whenSuccessful_returnsCreatedIssueNumbers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"new_issues":[{"id":"x1","number":5},{"id":"x2","number":6},{"id":"x3","number":7}]}`))
	}))
	defer srv.Close()

	handler := makeSplitIssue(testClient(srv.URL))
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":   "proj",
		"number": float64(1),
		"titles": "A,B,C",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !contains(text, "#5") || !contains(text, "#6") || !contains(text, "#7") {
		t.Errorf("expected new issue numbers in result, got: %s", text)
	}
}
