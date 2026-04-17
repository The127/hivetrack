package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHivemindConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/hivemind/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"grpc_url": "localhost:50051"})
	}))
	defer srv.Close()

	cfg, err := testClient(srv.URL).GetHivemindConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GrpcURL != "localhost:50051" {
		t.Errorf("expected localhost:50051, got %s", cfg.GrpcURL)
	}
}

func TestListDrones(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/drones" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "drone-1", "name": "worker", "project_slug": "proj", "status": "online"},
		})
	}))
	defer srv.Close()

	drones, err := testClient(srv.URL).ListDrones(context.Background(), "proj")
	if err != nil {
		t.Fatal(err)
	}
	if len(drones) != 1 || drones[0].ID != "drone-1" {
		t.Errorf("unexpected drones: %+v", drones)
	}
}

func TestGetDrone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/drones/drone-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "drone-1", "name": "worker", "project_slug": "proj"})
	}))
	defer srv.Close()

	drone, err := testClient(srv.URL).GetDrone(context.Background(), "proj", "drone-1")
	if err != nil {
		t.Fatal(err)
	}
	if drone.ID != "drone-1" {
		t.Errorf("expected drone-1, got %s", drone.ID)
	}
}

func TestCreateDroneToken(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/drones/tokens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]any{"token": "tok-abc"})
	}))
	defer srv.Close()

	result, err := testClient(srv.URL).CreateDroneToken(context.Background(), "proj", CreateDroneTokenRequest{
		Capabilities:   []string{"refinement"},
		MaxConcurrency: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != "tok-abc" {
		t.Errorf("expected tok-abc, got %s", result.Token)
	}
	caps, _ := gotBody["capabilities"].([]any)
	if len(caps) != 1 || caps[0] != "refinement" {
		t.Errorf("expected capabilities=[refinement], got %v", gotBody["capabilities"])
	}
}

func TestDeregisterDrone(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj/drones/drone-1/deregister" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeregisterDrone(context.Background(), "proj", "drone-1")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
}

func TestDeleteDrone(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/projects/proj/drones/drone-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteDrone(context.Background(), "proj", "drone-1")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
}

func TestRevokeDroneToken(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/projects/proj/drones/tokens/tok-abc" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).RevokeDroneToken(context.Background(), "proj", "tok-abc")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected API to be called")
	}
}
