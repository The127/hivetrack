package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/myproj" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": "p1", "slug": "myproj", "name": "My Project", "archetype": "software",
			"members": []map[string]any{{"user_id": "u1", "email": "a@b.com", "role": "project_admin"}},
		})
	}))
	defer srv.Close()

	p, err := testClient(srv.URL).GetProject(context.Background(), "myproj")
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "myproj" || len(p.Members) != 1 {
		t.Errorf("unexpected project: %+v", p)
	}
}

func TestCreateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": "new-proj-id"})
	}))
	defer srv.Close()

	id, err := testClient(srv.URL).CreateProject(context.Background(), CreateProjectRequest{
		Slug: "new", Name: "New Project", Archetype: "software",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "new-proj-id" {
		t.Errorf("expected new-proj-id, got %s", id)
	}
}

func TestUpdateProject(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).UpdateProject(context.Background(), "proj-id", UpdateProjectRequest{
		Name: Set("Renamed"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["name"] != "Renamed" {
		t.Errorf("expected name, got %v", gotBody["name"])
	}
}

func TestDeleteProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).DeleteProject(context.Background(), "proj-id")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddProjectMember(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).AddProjectMember(context.Background(), "proj", "user-id", ProjectRoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["role"] != "project_member" {
		t.Errorf("expected project_member, got %v", gotBody["role"])
	}
}

func TestRemoveProjectMember(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj/members/user-id" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := testClient(srv.URL).RemoveProjectMember(context.Background(), "proj", "user-id")
	if err != nil {
		t.Fatal(err)
	}
}
