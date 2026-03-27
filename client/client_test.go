package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(url string) *Client {
	return NewWithHTTPClient(url, func(_ context.Context) (string, error) {
		return "test-token", nil
	}, http.DefaultClient)
}

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"slug": "proj1", "name": "Project 1", "archetype": "software"},
			},
		})
	}))
	defer srv.Close()

	projects, err := testClient(srv.URL).ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 || projects[0].Slug != "proj1" {
		t.Errorf("unexpected projects: %+v", projects)
	}
}

func TestGetIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/myproj/issues/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "uuid-1", "number": 42, "title": "Test Issue",
			"type": "task", "status": "todo", "priority": "high",
			"estimate": "m", "triaged": true,
			"assignees": []any{}, "labels": []any{},
			"links": []any{
				map[string]any{"link_type": "blocks", "linked_issue_number": 10},
			},
			"checklist": []any{},
		})
	}))
	defer srv.Close()

	issue, err := testClient(srv.URL).GetIssue(context.Background(), "myproj", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Number != 42 || issue.Title != "Test Issue" {
		t.Errorf("unexpected issue: %+v", issue)
	}
	if len(issue.Links) != 1 || issue.Links[0].LinkType != LinkTypeBlocks {
		t.Errorf("unexpected links: %+v", issue.Links)
	}
}

func TestTriageIssue(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	priority := "high"
	estimate := "m"
	err := testClient(srv.URL).TriageIssue(context.Background(), "proj", 5, TriageIssueRequest{
		Status:   "todo",
		Priority: &priority,
		Estimate: &estimate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["priority"] != "high" || gotBody["estimate"] != "m" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"code":"not_found"}]}`))
	}))
	defer srv.Close()

	_, err := testClient(srv.URL).GetIssue(context.Background(), "proj", 999)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestResolveLabelNames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"labels": []map[string]any{
				{"id": "uuid-1", "name": "groomed", "color": "#00ff00"},
				{"id": "uuid-2", "name": "bug", "color": "#ff0000"},
			},
		})
	}))
	defer srv.Close()

	ids, err := testClient(srv.URL).ResolveLabelNames(context.Background(), "proj", "groomed, Bug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "uuid-1" || ids[1] != "uuid-2" {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestBatchUpdateIssues(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/issues/batch-update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"Updated": 3})
	}))
	defer srv.Close()

	result, err := testClient(srv.URL).BatchUpdateIssues(context.Background(), "proj", BatchUpdateIssuesRequest{
		Numbers: []int{1, 2, 3},
		Status:  Set("in_progress"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated != 3 {
		t.Errorf("expected 3 updated, got %d", result.Updated)
	}
}
