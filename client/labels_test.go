package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListLabels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"labels": []map[string]any{{"id": "l1", "name": "bug", "color": "#ff0000"}},
		})
	}))
	defer srv.Close()

	labels, err := testClient(srv.URL).ListLabels(context.Background(), "proj")
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 || labels[0].Name != "bug" {
		t.Errorf("unexpected: %+v", labels)
	}
}

func TestCreateLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ID": "new-label-id"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).CreateLabel(context.Background(), "proj", "feature", "#00ff00")
	if err != nil {
		t.Fatal(err)
	}
	if id != "new-label-id" {
		t.Errorf("expected new-label-id, got %s", id)
	}
}

func TestUpdateLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	name := "renamed"
	err := testClient(srv.URL).UpdateLabel(context.Background(), "proj", "l1", UpdateLabelRequest{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteLabel(context.Background(), "proj", "l1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolveLabelNames_notFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"labels": []map[string]any{}})
	}))
	defer srv.Close()

	_, err := testClient(srv.URL).ResolveLabelNames(context.Background(), "proj", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown label")
	}
}
