package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSprints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"sprints": []map[string]any{{"id": "s1", "name": "Sprint 1", "status": "active"}},
		})
	}))
	defer srv.Close()

	sprints, err := testClient(srv.URL).ListSprints(context.Background(), "proj")
	if err != nil {
		t.Fatal(err)
	}
	if len(sprints) != 1 || sprints[0].Name != "Sprint 1" {
		t.Errorf("unexpected sprints: %+v", sprints)
	}
}

func TestCreateSprint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "sprint-uuid"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).CreateSprint(context.Background(), "proj", CreateSprintRequest{Name: "Sprint 1"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "sprint-uuid" {
		t.Errorf("expected sprint-uuid, got %s", id)
	}
}

func TestUpdateSprint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).UpdateSprint(context.Background(), "proj", "s1", UpdateSprintRequest{
		Status: Set("completed"),
		Force:  Set(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["force"] != true {
		t.Errorf("expected force=true, got %v", gotBody["force"])
	}
}

func TestDeleteSprint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteSprint(context.Background(), "proj", "s1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSprintBurndown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"total": 10, "start_remaining": 10, "end_remaining": 3,
			"points": []map[string]any{{"date": "2026-01-01", "remaining": 8}},
		})
	}))
	defer srv.Close()

	bd, err := testClient(srv.URL).GetSprintBurndown(context.Background(), "proj", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if bd.Total != 10 || len(bd.Points) != 1 {
		t.Errorf("unexpected burndown: %+v", bd)
	}
}
