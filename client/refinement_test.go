package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartRefinementSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/issues/5/refinement/start" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"SessionID": "sess-uuid"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).StartRefinementSession(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if id != "sess-uuid" {
		t.Errorf("expected sess-uuid, got %s", id)
	}
}

func TestSendRefinementMessage(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/issues/5/refinement/message" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).SendRefinementMessage(context.Background(), "proj", 5, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["content"] != "hello" {
		t.Errorf("expected content=hello, got %v", gotBody["content"])
	}
}

func TestGetRefinementSession_returnsSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/issues/5/refinement/session" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":               "sess-uuid",
			"issue_id":         "issue-uuid",
			"status":           "active",
			"current_phase":    "actor_goal",
			"messages":         []any{},
			"partial_response": "",
			"is_generating":    false,
		})
	}))
	defer srv.Close()

	session, err := testClient(srv.URL).GetRefinementSession(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.ID != "sess-uuid" {
		t.Errorf("expected sess-uuid, got %s", session.ID)
	}
	if session.Status != RefinementSessionActive {
		t.Errorf("expected active, got %s", session.Status)
	}
}

func TestGetRefinementSession_returnsNilWhenNoSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("null"))
	}))
	defer srv.Close()

	session, err := testClient(srv.URL).GetRefinementSession(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if session != nil {
		t.Errorf("expected nil session, got %+v", session)
	}
}

func TestAcceptRefinementProposal(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/issues/5/refinement/accept" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).AcceptRefinementProposal(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
}

func TestAdvanceRefinementPhase_nextPhase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/issues/5/refinement/advance-phase" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"Phase": "main_scenario"})
	}))
	defer srv.Close()

	phase, err := testClient(srv.URL).AdvanceRefinementPhase(context.Background(), "proj", 5, "")
	if err != nil {
		t.Fatal(err)
	}
	if phase != "main_scenario" {
		t.Errorf("expected main_scenario, got %s", phase)
	}
}

func TestAdvanceRefinementPhase_targetPhase(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]any{"Phase": "bdd_scenarios"})
	}))
	defer srv.Close()

	phase, err := testClient(srv.URL).AdvanceRefinementPhase(context.Background(), "proj", 5, "bdd_scenarios")
	if err != nil {
		t.Fatal(err)
	}
	if phase != "bdd_scenarios" {
		t.Errorf("expected bdd_scenarios, got %s", phase)
	}
	if gotBody["target_phase"] != "bdd_scenarios" {
		t.Errorf("expected target_phase in body, got %v", gotBody["target_phase"])
	}
}
