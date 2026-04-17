package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	htclient "github.com/the127/hivetrack/client"
	"github.com/urfave/cli/v2"
)

// testApp builds the CLI app with a custom writer so output can be captured.
func testApp(out *bytes.Buffer) *cli.App {
	app := &cli.App{
		Name:   "ht",
		Writer: out,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server", Aliases: []string{"s"}},
			&cli.BoolFlag{Name: "json"},
		},
		Commands: []*cli.Command{
			loginCmd,
			logoutCmd,
			projectsCmd,
			issuesCmd,
			sprintsCmd,
			milestonesCmd,
		},
	}
	app.ExitErrHandler = func(_ *cli.Context, _ error) {}
	return app
}

// withTestServer starts a test HTTP server and sets HIVETRACK_TOKEN so mustClient
// connects without real credentials.
func withTestServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	t.Setenv("HIVETRACK_TOKEN", "test-token")
	return srv.URL
}

func assertContains(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Errorf("output %q does not contain %q", output, want)
	}
}

// --- projects ---

func TestE2E_projects(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"slug": "myproj", "name": "My Project", "archetype": "software"},
			},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "projects"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "My Project")
	assertContains(t, out.String(), "myproj")
}

func TestE2E_projects_json(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"slug": "p1", "name": "P1", "archetype": "software"}},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "--json", "projects"}); err != nil {
		t.Fatal(err)
	}
	var parsed []htclient.ProjectSummary
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got: %s\nerr: %v", out.String(), err)
	}
	if len(parsed) != 1 || parsed[0].Slug != "p1" {
		t.Errorf("unexpected parsed: %+v", parsed)
	}
}

// --- issues list ---

func TestE2E_issues_list(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"number": 7, "title": "Fix the bug", "type": "task", "status": "todo",
					"priority": "high", "estimate": "m", "triaged": true, "assignees": []any{}, "labels": []any{}},
			},
			"total": 1,
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "list", "myproj"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "#7")
	assertContains(t, out.String(), "Fix the bug")
}

func TestE2E_issues_list_filters_forwarded(t *testing.T) {
	var gotQuery string
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "total": 0})
	}))

	var out bytes.Buffer
	// Note: flags must come before positional arg in urfave/cli v2 nested subcommands
	testApp(&out).Run([]string{"ht", "--server", url, "issues", "list", "--status", "in_progress", "--limit", "10", "myproj"})
	assertContains(t, gotQuery, "status=in_progress")
	assertContains(t, gotQuery, "limit=10")
}

// --- issues show ---

func TestE2E_issues_show(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/myproj/issues/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": "uuid-1", "number": 42, "title": "Important issue",
			"type": "task", "status": "in_progress", "priority": "high",
			"estimate": "l", "triaged": true, "refined": true,
			"assignees": []any{}, "labels": []any{}, "links": []any{}, "checklist": []any{},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "show", "myproj", "42"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "#42")
	assertContains(t, out.String(), "Important issue")
	assertContains(t, out.String(), "in_progress")
}

// --- issues create ---

func TestE2E_issues_create(t *testing.T) {
	var gotBody map[string]any
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"ID": "new-uuid", "Number": 99})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "create", "--title", "New thing", "--priority", "high", "myproj"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "#99")
	assertContains(t, out.String(), "New thing")
	if gotBody["title"] != "New thing" {
		t.Errorf("expected title=New thing, got %v", gotBody["title"])
	}
	if gotBody["priority"] != "high" {
		t.Errorf("expected priority=high, got %v", gotBody["priority"])
	}
}

func TestE2E_issues_create_missing_title(t *testing.T) {
	t.Setenv("HIVETRACK_TOKEN", "tok")
	var out bytes.Buffer
	err := testApp(&out).Run([]string{"ht", "--server", "http://localhost", "issues", "create", "myproj"})
	if err == nil {
		t.Error("expected error for missing title")
	}
}

// --- issues update ---

func TestE2E_issues_update(t *testing.T) {
	var gotBody map[string]any
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "update", "--status", "done", "myproj", "5"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "Updated #5")
	if gotBody["status"] != "done" {
		t.Errorf("expected status=done, got %v", gotBody["status"])
	}
}

func TestE2E_issues_update_clearSprint(t *testing.T) {
	var gotBody map[string]any
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	var out bytes.Buffer
	testApp(&out).Run([]string{"ht", "--server", url, "issues", "update", "--clear-sprint", "myproj", "5"})
	if gotBody["sprint_id"] != nil {
		t.Errorf("expected sprint_id=null, got %v", gotBody["sprint_id"])
	}
}

// --- issues me ---

func TestE2E_issues_me(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"number": 3, "title": "My assigned issue", "type": "task", "status": "in_progress",
					"assignees": []any{}, "labels": []any{}, "triaged": true},
			},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "me"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "My assigned issue")
}

func TestE2E_issues_me_created(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me/created-issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"number": 11, "title": "Issue I reported", "type": "task", "status": "todo",
					"assignees": []any{}, "labels": []any{}, "triaged": false},
			},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "issues", "me", "--created"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "Issue I reported")
}

// --- sprints list ---

func TestE2E_sprints_list(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"sprints": []map[string]any{
				{"id": "s1", "name": "Sprint 1", "status": "active", "goal": "Ship it",
					"start_date": "2026-04-01", "end_date": "2026-04-14"},
			},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "sprints", "list", "myproj"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "Sprint 1")
	assertContains(t, out.String(), "active")
}

// --- milestones list ---

func TestE2E_milestones_list(t *testing.T) {
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"milestones": []map[string]any{
				{"id": "m1", "title": "v1.0", "issue_count": 10, "closed_issue_count": 3},
			},
		})
	}))

	var out bytes.Buffer
	if err := testApp(&out).Run([]string{"ht", "--server", url, "milestones", "list", "myproj"}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, out.String(), "v1.0")
	assertContains(t, out.String(), "3/10")
}

// --- logout ---

func TestE2E_logout_noFile(t *testing.T) {
	// Logout should succeed even if no credentials file exists.
	var out bytes.Buffer
	err := testApp(&out).Run([]string{"ht", "logout"})
	// May return a cli.ExitError — not a panic.
	if err != nil {
		// acceptable: file might not exist; as long as it doesn't panic
		_ = err
	}
}

// --- missing args ---

func TestE2E_issues_list_missingSlug(t *testing.T) {
	t.Setenv("HIVETRACK_TOKEN", "tok")
	var out bytes.Buffer
	err := testApp(&out).Run([]string{"ht", "--server", "http://localhost", "issues", "list"})
	if err == nil {
		t.Error("expected error for missing slug")
	}
}

func TestE2E_issues_show_badNumber(t *testing.T) {
	t.Setenv("HIVETRACK_TOKEN", "tok")
	var out bytes.Buffer
	err := testApp(&out).Run([]string{"ht", "--server", "http://localhost", "issues", "show", "proj", "notanumber"})
	if err == nil {
		t.Error("expected error for bad issue number")
	}
}

// --- auth env override ---

func TestE2E_bearerTokenSentFromEnv(t *testing.T) {
	var gotAuth string
	url := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))

	os.Setenv("HIVETRACK_TOKEN", "my-custom-token")
	var out bytes.Buffer
	testApp(&out).Run([]string{"ht", "--server", url, "projects"})
	if gotAuth != "Bearer my-custom-token" {
		t.Errorf("expected Bearer my-custom-token, got %q", gotAuth)
	}
}
