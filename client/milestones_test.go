package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListMilestones(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"milestones": []map[string]any{{"id": "m1", "title": "v1.0", "issue_count": 5, "closed_issue_count": 2}},
		})
	}))
	defer srv.Close()

	ms, err := testClient(srv.URL).ListMilestones(context.Background(), "proj")
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 || ms[0].Title != "v1.0" {
		t.Errorf("unexpected: %+v", ms)
	}
}

func TestCreateMilestone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "ms-uuid"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).CreateMilestone(context.Background(), "proj", CreateMilestoneRequest{Title: "v2.0"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "ms-uuid" {
		t.Errorf("expected ms-uuid, got %s", id)
	}
}

func TestUpdateMilestone(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).UpdateMilestone(context.Background(), "proj", "m1", UpdateMilestoneRequest{
		Close: Set(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["close"] != true {
		t.Errorf("expected close=true, got %v", gotBody["close"])
	}
}

func TestDeleteMilestone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteMilestone(context.Background(), "proj", "m1")
	if err != nil {
		t.Fatal(err)
	}
}
