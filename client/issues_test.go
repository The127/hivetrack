package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIssues_withFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "todo" {
			t.Errorf("expected status=todo, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("label_id") != "lbl-uuid" {
			t.Errorf("expected label_id=lbl-uuid, got %s", r.URL.Query().Get("label_id"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"number": 1, "title": "Test", "status": "todo", "type": "task"}},
			"total": 1,
		})
	}))
	defer srv.Close()

	items, total, err := testClient(srv.URL).ListIssues(context.Background(), "proj", ListIssuesOptions{
		Status:  "todo",
		LabelID: "lbl-uuid",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d (total %d)", len(items), total)
	}
}

func TestCreateIssue(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"ID": "new-uuid", "Number": 42})
	}))
	defer srv.Close()

	result, err := testClient(srv.URL).CreateIssue(context.Background(), "proj", CreateIssueRequest{
		Title:    "New Issue",
		Type:     "task",
		Priority: "high",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Number != 42 {
		t.Errorf("expected number 42, got %d", result.Number)
	}
	if gotBody["title"] != "New Issue" {
		t.Errorf("expected title in body, got %v", gotBody["title"])
	}
}

func TestUpdateIssue_sendsCorrectBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).UpdateIssue(context.Background(), "proj", 5, UpdateIssueRequest{
		Title:         ptr("Updated"),
		ClearSprintID: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["title"] != "Updated" {
		t.Errorf("expected title, got %v", gotBody["title"])
	}
	if gotBody["sprint_id"] != nil {
		t.Errorf("expected sprint_id=null, got %v", gotBody["sprint_id"])
	}
}

func TestDeleteIssue(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteIssue(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
}

func TestRefineIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/issues/3/refine" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).RefineIssue(context.Background(), "proj", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSplitIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"new_issues": []map[string]any{{"id": "a", "number": 10}, {"id": "b", "number": 11}},
		})
	}))
	defer srv.Close()

	result, err := testClient(srv.URL).SplitIssue(context.Background(), "proj", 5, []string{"Part A", "Part B"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewIssues) != 2 {
		t.Errorf("expected 2 new issues, got %d", len(result.NewIssues))
	}
}

func TestAddIssueLink(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).AddIssueLink(context.Background(), "proj", 1, LinkTypeBlocks, 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["link_type"] != "blocks" {
		t.Errorf("expected blocks, got %v", gotBody["link_type"])
	}
}

func TestAddChecklistItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ID": "item-uuid"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).AddChecklistItem(context.Background(), "proj", 1, "Do something")
	if err != nil {
		t.Fatal(err)
	}
	if id != "item-uuid" {
		t.Errorf("expected item-uuid, got %s", id)
	}
}

func TestUpdateChecklistItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	done := true
	err := testClient(srv.URL).UpdateChecklistItem(context.Background(), "proj", 1, "item-id", UpdateChecklistItemRequest{Done: &done})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveChecklistItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).RemoveChecklistItem(context.Background(), "proj", 1, "item-id")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMyIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"number": 1, "title": "My Issue", "type": "task", "status": "todo"}},
		})
	}))
	defer srv.Close()

	items, err := testClient(srv.URL).GetMyIssues(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}
