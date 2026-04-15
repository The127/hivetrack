//go:build integration

// Package smoke contains HTTP-level smoke tests.
// They spin up a full server against a real database and exercise key endpoints.
//
// Run: just test-integration (requires HIVETRACK_DATABASE_URL)
package smoke_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/The127/ioc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/database"
	"github.com/the127/hivetrack/internal/server"
	"github.com/the127/hivetrack/internal/setup"
)

const smokeToken = "smoke-test-token-do-not-use-in-prod"
const smokeUserEmail = "smoke@test.internal"

var testServer *httptest.Server

func TestMain(m *testing.M) {
	dsn := os.Getenv("HIVETRACK_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "HIVETRACK_DATABASE_URL not set — skipping smoke tests")
		os.Exit(0)
	}

	db, err := database.Open(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{AllowedOrigins: []string{"*"}},
		MCP: config.MCPConfig{
			APIToken:  smokeToken,
			UserEmail: smokeUserEmail,
		},
		// OIDC authority is intentionally blank — MCP token path bypasses OIDC.
	}

	dc := ioc.NewDependencyCollection()
	setup.Database(dc, db)
	setup.Services(dc, cfg)
	broker := setup.Events(dc)
	setup.Mediator(dc, broker)
	dp := dc.BuildProvider()

	testServer = httptest.NewServer(server.New(dp))
	defer testServer.Close()

	os.Exit(m.Run())
}

// do performs an authenticated HTTP request against the test server.
func do(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, testServer.URL+path, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+smokeToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, r *http.Response, dst any) {
	t.Helper()
	defer r.Body.Close()
	require.NoError(t, json.NewDecoder(r.Body).Decode(dst))
}

// TestSmoke_OIDCConfig verifies the public auth config endpoint.
func TestSmoke_OIDCConfig(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/api/v1/auth/oidc-config")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestSmoke_UnauthorizedWithoutToken verifies that protected routes require auth.
func TestSmoke_UnauthorizedWithoutToken(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/api/v1/projects")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestSmoke_ProjectCRUD exercises create, get, and list for projects.
func TestSmoke_ProjectCRUD(t *testing.T) {
	slug := fmt.Sprintf("smoke-%s", uuid.New().String()[:8])

	// Create project
	createResp := do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"slug":      slug,
		"name":      "Smoke Test Project",
		"archetype": "software",
	})
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	decodeJSON(t, createResp, &created)
	require.NotEmpty(t, created.ID)

	// Get project
	getResp := do(t, http.MethodGet, "/api/v1/projects/"+slug, nil)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	var got struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	decodeJSON(t, getResp, &got)
	assert.Equal(t, slug, got.Slug)
	assert.Equal(t, "Smoke Test Project", got.Name)

	// List projects — created project must appear
	listResp := do(t, http.MethodGet, "/api/v1/projects", nil)
	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	var list struct {
		Items []struct {
			Slug string `json:"slug"`
		} `json:"items"`
	}
	decodeJSON(t, listResp, &list)
	var found bool
	for _, p := range list.Items {
		if p.Slug == slug {
			found = true
			break
		}
	}
	assert.True(t, found, "created project should appear in list")
}

// TestSmoke_IssueCRUD exercises create, get, and list for issues.
func TestSmoke_IssueCRUD(t *testing.T) {
	slug := fmt.Sprintf("smoke-%s", uuid.New().String()[:8])

	// Create project first
	createProjResp := do(t, http.MethodPost, "/api/v1/projects", map[string]any{
		"slug":      slug,
		"name":      "Issue Smoke Project",
		"archetype": "software",
	})
	require.Equal(t, http.StatusCreated, createProjResp.StatusCode)
	createProjResp.Body.Close()

	// Create issue
	createIssueResp := do(t, http.MethodPost, "/api/v1/projects/"+slug+"/issues", map[string]any{
		"title": "Smoke test issue",
		"type":  "task",
	})
	assert.Equal(t, http.StatusCreated, createIssueResp.StatusCode)
	var createdIssue struct {
		Number int `json:"number"`
	}
	decodeJSON(t, createIssueResp, &createdIssue)
	require.Greater(t, createdIssue.Number, 0)

	// Get issue
	getIssueResp := do(t, http.MethodGet,
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, createdIssue.Number), nil)
	assert.Equal(t, http.StatusOK, getIssueResp.StatusCode)
	var issue struct {
		Title string `json:"title"`
	}
	decodeJSON(t, getIssueResp, &issue)
	assert.Equal(t, "Smoke test issue", issue.Title)

	// List issues
	listResp := do(t, http.MethodGet, "/api/v1/projects/"+slug+"/issues", nil)
	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	var list struct {
		Total int `json:"total"`
	}
	decodeJSON(t, listResp, &list)
	assert.Equal(t, 1, list.Total)
}

// TestSmoke_GetMyIssues verifies the /me/issues endpoint responds correctly.
func TestSmoke_GetMyIssues(t *testing.T) {
	resp := do(t, http.MethodGet, "/api/v1/me/issues", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestSmoke_NotFound verifies 404 for unknown project slugs.
func TestSmoke_NotFound(t *testing.T) {
	resp := do(t, http.MethodGet, "/api/v1/projects/no-such-project-"+uuid.New().String(), nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
