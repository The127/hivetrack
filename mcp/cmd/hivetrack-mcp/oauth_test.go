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

func TestHandleRegister_InvalidJSON(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
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

func TestHandleCallback_UnknownState(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/callback?code=abc&state=bogus", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_request" {
		t.Fatalf("expected invalid_request, got %s", resp["error"])
	}
}

func TestHandleCallback_MissingCode(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/callback?state=something", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCallback_ErrorFromProvider(t *testing.T) {
	p := newTestProxy()
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/oauth/callback?error=access_denied&error_description=user+denied", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "access_denied" {
		t.Fatalf("expected access_denied, got %s", resp["error"])
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
		createdAt:           time.Now(),
	})

	body := "grant_type=authorization_code&code=proxy-code&code_verifier=" + verifier
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
		createdAt:           time.Now(),
	})

	body := "grant_type=authorization_code&code=code-1&code_verifier=wrong-verifier"
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
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
		createdAt:           time.Now().Add(-10 * time.Minute),
	})

	body := "grant_type=authorization_code&code=old-code&code_verifier=" + verifier
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
