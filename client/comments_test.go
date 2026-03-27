package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListComments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"id": "c1", "body": "Hello", "author_name": "Alice"}},
			"total": 1,
		})
	}))
	defer srv.Close()

	comments, total, err := testClient(srv.URL).ListComments(context.Background(), "proj", 1)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(comments) != 1 || comments[0].Body != "Hello" {
		t.Errorf("unexpected: %+v", comments)
	}
}

func TestCreateComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).CreateComment(context.Background(), "proj", 1, "A comment")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).UpdateComment(context.Background(), "proj", 1, "c-id", "Updated body")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteComment(context.Background(), "proj", 1, "c-id")
	if err != nil {
		t.Fatal(err)
	}
}
