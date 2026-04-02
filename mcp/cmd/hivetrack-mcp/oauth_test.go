package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestProxy() *oauthProxy {
	return &oauthProxy{
		apiURL:      "http://localhost:8086",
		externalURL: "http://localhost:8080",
	}
}

func TestHandleMetadata(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var meta map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &meta); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checks := map[string]string{
		"issuer":                 "http://localhost:8080",
		"authorization_endpoint": "http://localhost:8080/oauth/authorize",
		"token_endpoint":         "http://localhost:8080/oauth/token",
		"registration_endpoint":  "http://localhost:8080/oauth/register",
	}
	for key, want := range checks {
		got, _ := meta[key].(string)
		if got != want {
			t.Errorf("%s: got %q, want %q", key, got, want)
		}
	}
}

func TestHandleRegister(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	body := `{"redirect_uris":["http://127.0.0.1:9999/callback"],"client_name":"test-client"}`
	req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	clientID, ok := resp["client_id"].(string)
	if !ok || clientID == "" {
		t.Fatal("expected non-empty client_id")
	}

	// Verify client is stored.
	if _, ok := p.clients.Load(clientID); !ok {
		t.Fatal("client should be stored in memory")
	}
}

func TestHandleAuthorize_MissingParams(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAuthorize_UnknownClient(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=bogus&redirect_uri=http://x&code_challenge=abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_client" {
		t.Fatalf("expected invalid_client error, got %s", resp["error"])
	}
}

func TestHandlePoll_UnknownFlow(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/poll?id=nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlePoll_Pending(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	pa := &pendingAuth{createdAt: time.Now()}
	p.pending.Store("flow-1", pa)

	req := httptest.NewRequest("GET", "/oauth/poll?id=flow-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "pending" {
		t.Fatalf("expected pending, got %s", resp["status"])
	}
}

func TestHandlePoll_Complete(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	pa := &pendingAuth{
		redirectURI: "http://127.0.0.1:9999/callback",
		state:       "original-state",
		createdAt:   time.Now(),
		completed:   true,
		proxyCode:   "proxy-code-123",
	}
	p.pending.Store("flow-2", pa)

	req := httptest.NewRequest("GET", "/oauth/poll?id=flow-2", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "complete" {
		t.Fatalf("expected complete, got %s", resp["status"])
	}

	redirect, _ := resp["redirect"].(string)
	if !strings.Contains(redirect, "code=proxy-code-123") {
		t.Fatalf("redirect should contain proxy code: %s", redirect)
	}
	if !strings.Contains(redirect, "state=original-state") {
		t.Fatalf("redirect should contain original state: %s", redirect)
	}

	// Flow should be cleaned up.
	if _, ok := p.pending.Load("flow-2"); ok {
		t.Fatal("pending flow should be deleted after poll returns complete")
	}
}

func makePKCE(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func TestHandleToken_AuthCode_Success(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	verifier := "test-verifier-that-is-long-enough-for-pkce"
	challenge := makePKCE(verifier)

	p.codes.Store("proxy-code", &proxyCodeEntry{
		accessToken:         "real-access-token",
		refreshToken:        "real-refresh-token",
		expiresIn:           3600,
		codeChallenge:       challenge,
		codeChallengeMethod: "S256",
		clientID:            "test-client",
		redirectURI:         "http://127.0.0.1:9999/callback",
		createdAt:           time.Now(),
	})

	body := "grant_type=authorization_code&code=proxy-code&code_verifier=" + verifier + "&client_id=test-client"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["access_token"] != "real-access-token" {
		t.Fatalf("expected real-access-token, got %v", resp["access_token"])
	}
	if resp["refresh_token"] != "real-refresh-token" {
		t.Fatalf("expected real-refresh-token, got %v", resp["refresh_token"])
	}
	if resp["token_type"] != "Bearer" {
		t.Fatalf("expected Bearer, got %v", resp["token_type"])
	}
}

func TestHandleToken_AuthCode_ReusedCode(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	// Code doesn't exist (already consumed or never existed).
	body := "grant_type=authorization_code&code=nonexistent&code_verifier=whatever"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %s", resp["error"])
	}
}

func TestHandleToken_AuthCode_WrongVerifier(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	challenge := makePKCE("correct-verifier")
	p.codes.Store("code-1", &proxyCodeEntry{
		accessToken:         "token",
		codeChallenge:       challenge,
		codeChallengeMethod: "S256",
		clientID:            "client",
		createdAt:           time.Now(),
	})

	body := "grant_type=authorization_code&code=code-1&code_verifier=wrong-verifier&client_id=client"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %s", resp["error"])
	}

	// Code should be consumed even on PKCE failure.
	if _, ok := p.codes.Load("code-1"); ok {
		t.Fatal("code should be consumed after failed exchange")
	}
}

func TestHandleToken_AuthCode_ExpiredCode(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	verifier := "test-verifier"
	challenge := makePKCE(verifier)
	p.codes.Store("old-code", &proxyCodeEntry{
		accessToken:         "token",
		codeChallenge:       challenge,
		codeChallengeMethod: "S256",
		clientID:            "client",
		createdAt:           time.Now().Add(-10 * time.Minute), // expired
	})

	body := "grant_type=authorization_code&code=old-code&code_verifier=" + verifier + "&client_id=client"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleToken_ClientMismatch(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	verifier := "test-verifier"
	challenge := makePKCE(verifier)
	p.codes.Store("code-2", &proxyCodeEntry{
		accessToken:         "token",
		codeChallenge:       challenge,
		codeChallengeMethod: "S256",
		clientID:            "real-client",
		createdAt:           time.Now(),
	})

	body := "grant_type=authorization_code&code=code-2&code_verifier=" + verifier + "&client_id=wrong-client"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleToken_UnsupportedGrantType(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	body := "grant_type=client_credentials"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "unsupported_grant_type" {
		t.Fatalf("expected unsupported_grant_type, got %s", resp["error"])
	}
}

func TestVerifyPKCE(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	if !verifyPKCE(verifier, challenge, "S256") {
		t.Fatal("valid PKCE should pass")
	}
	if verifyPKCE("wrong-verifier", challenge, "S256") {
		t.Fatal("wrong verifier should fail")
	}
	if verifyPKCE(verifier, challenge, "plain") {
		t.Fatal("unsupported method should fail")
	}
}
