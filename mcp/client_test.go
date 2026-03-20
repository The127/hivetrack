package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// staticProvider is a TokenProvider that always returns the same token.
type staticProvider struct{ tc tokenCache }

func (s staticProvider) ProvideToken(_ context.Context) (tokenCache, error) { return s.tc, nil }

func freshProvider(token string) TokenProvider {
	return staticProvider{tc: tokenCache{AccessToken: token, Expiry: time.Now().Add(time.Hour)}}
}

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("expected /api/v1/projects, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, freshProvider("test-token"))
	data, err := client.get("/api/v1/projects", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["items"]; !ok {
		t.Error("expected items key in response")
	}
}

func TestClient_GetWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "todo" {
			t.Errorf("expected status=todo, got %s", r.URL.Query().Get("status"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[],"total":0}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, freshProvider("test-token"))
	q := url.Values{}
	q.Set("status", "todo")
	_, err := client.get("/api/v1/projects/my-project/issues", q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["title"] != "Test issue" {
			t.Errorf("expected title 'Test issue', got %v", body["title"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"abc-123","number":1}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, freshProvider("test-token"))
	data, err := client.post("/api/v1/projects/my-project/issues", map[string]any{
		"title": "Test issue",
		"type":  "task",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(data, &result)
	if result["number"] != float64(1) {
		t.Errorf("expected number 1, got %v", result["number"])
	}
}

func TestClient_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, freshProvider("test-token"))
	data, err := client.patch("/api/v1/projects/my-project/issues/1", map[string]any{
		"status": "done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 204 returns {"ok":true}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["ok"] != true {
		t.Errorf("expected ok:true for 204 response")
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"code":"not_found","message":"project not found"}]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, freshProvider("test-token"))
	_, err := client.get("/api/v1/projects/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
