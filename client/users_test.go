package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "u1", "email": "a@b.com", "is_admin": true})
	}))
	defer srv.Close()

	user, err := testClient(srv.URL).GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "a@b.com" || !user.IsAdmin {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestListUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"users": []map[string]any{{"id": "u1", "email": "a@b.com"}, {"id": "u2", "email": "c@d.com"}},
		})
	}))
	defer srv.Close()

	users, err := testClient(srv.URL).ListUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}
