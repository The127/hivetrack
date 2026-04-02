package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	htclient "github.com/the127/hivetrack/client"
)

// oauthProxy implements an OAuth 2.1 Authorization Server facade.
// MCP clients (Claude Code) discover this via /.well-known/oauth-authorization-server,
// register dynamically, and go through an authorization code + PKCE flow.
// Behind the scenes, /oauth/authorize uses OIDC device flow with the real provider.
type oauthProxy struct {
	apiURL      string // Hivetrack API URL (for device flow)
	externalURL string // This server's public URL (for metadata endpoints)

	clients sync.Map // client_id (string) -> *clientRegistration
	pending sync.Map // flow ID (string) -> *pendingAuth
	codes   sync.Map // proxy code (string) -> *proxyCodeEntry
}

type clientRegistration struct {
	clientID     string
	redirectURIs []string
	createdAt    time.Time
}

type pendingAuth struct {
	clientID            string
	redirectURI         string
	state               string
	codeChallenge       string
	codeChallengeMethod string
	scope               string
	createdAt           time.Time

	// Set by the background goroutine once device flow completes.
	mu        sync.Mutex
	completed bool
	proxyCode string
}

type proxyCodeEntry struct {
	accessToken  string
	refreshToken string
	expiresIn    int

	codeChallenge       string
	codeChallengeMethod string
	clientID            string
	redirectURI         string
	createdAt           time.Time
}

func newOAuthProxy(apiURL, externalURL string) *oauthProxy {
	p := &oauthProxy{
		apiURL:      apiURL,
		externalURL: externalURL,
	}
	go p.cleanup()
	return p
}

// RegisterRoutes adds the OAuth proxy endpoints to the given mux.
func (p *oauthProxy) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", p.handleMetadata)
	mux.HandleFunc("POST /oauth/register", p.handleRegister)
	mux.HandleFunc("GET /oauth/authorize", p.handleAuthorize)
	mux.HandleFunc("GET /oauth/poll", p.handlePoll)
	mux.HandleFunc("POST /oauth/token", p.handleToken)
}

func (p *oauthProxy) handleMetadata(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"issuer":                                p.externalURL,
		"authorization_endpoint":                p.externalURL + "/oauth/authorize",
		"token_endpoint":                        p.externalURL + "/oauth/token",
		"registration_endpoint":                 p.externalURL + "/oauth/register",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	})
}

func (p *oauthProxy) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RedirectURIs []string `json:"redirect_uris"`
		ClientName   string   `json:"client_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	clientID := randomHex(16)
	reg := &clientRegistration{
		clientID:     clientID,
		redirectURIs: body.RedirectURIs,
		createdAt:    time.Now(),
	}
	p.clients.Store(clientID, reg)

	respondJSON(w, http.StatusCreated, map[string]any{
		"client_id":                    clientID,
		"client_id_issued_at":          time.Now().Unix(),
		"redirect_uris":               body.RedirectURIs,
		"client_name":                  body.ClientName,
		"token_endpoint_auth_method":   "none",
		"grant_types":                  []string{"authorization_code", "refresh_token"},
		"response_types":              []string{"code"},
	})
}

func (p *oauthProxy) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	scope := q.Get("scope")

	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing required parameters")
		return
	}

	if _, ok := p.clients.Load(clientID); !ok {
		respondOAuthError(w, http.StatusBadRequest, "invalid_client", "unknown client_id")
		return
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	// Start OIDC device flow with the real provider.
	flow, err := htclient.InitDeviceFlow(p.apiURL)
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "failed to start device flow: "+err.Error())
		return
	}

	authURL := flow.VerificationURIComplete
	if authURL == "" {
		authURL = flow.VerificationURI
	}

	flowID := randomHex(16)
	pa := &pendingAuth{
		clientID:            clientID,
		redirectURI:         redirectURI,
		state:               state,
		codeChallenge:       codeChallenge,
		codeChallengeMethod: codeChallengeMethod,
		scope:               scope,
		createdAt:           time.Now(),
	}
	p.pending.Store(flowID, pa)

	// Poll for token in background.
	go p.waitForDeviceToken(flowID, pa, flow)

	pollURL := p.externalURL + "/oauth/poll?id=" + flowID

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, authPageHTML, authURL, authURL, pollURL)
}

func (p *oauthProxy) waitForDeviceToken(flowID string, pa *pendingAuth, flow *htclient.DeviceFlow) {
	tc, err := flow.WaitForToken(context.Background())
	if err != nil {
		// Flow failed — remove the pending entry so poll returns an error.
		p.pending.Delete(flowID)
		return
	}

	proxyCode := randomHex(24)
	p.codes.Store(proxyCode, &proxyCodeEntry{
		accessToken:         tc.AccessToken,
		refreshToken:        tc.RefreshToken,
		expiresIn:           int(time.Until(tc.Expiry).Seconds()),
		codeChallenge:       pa.codeChallenge,
		codeChallengeMethod: pa.codeChallengeMethod,
		clientID:            pa.clientID,
		redirectURI:         pa.redirectURI,
		createdAt:           time.Now(),
	})

	pa.mu.Lock()
	pa.completed = true
	pa.proxyCode = proxyCode
	pa.mu.Unlock()
}

func (p *oauthProxy) handlePoll(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("id")
	if flowID == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing id parameter")
		return
	}

	val, ok := p.pending.Load(flowID)
	if !ok {
		respondOAuthError(w, http.StatusNotFound, "invalid_request", "unknown or expired flow")
		return
	}

	pa := val.(*pendingAuth)
	pa.mu.Lock()
	completed := pa.completed
	proxyCode := pa.proxyCode
	pa.mu.Unlock()

	if !completed {
		respondJSON(w, http.StatusOK, map[string]string{"status": "pending"})
		return
	}

	// Build the redirect URL with the proxy code.
	redirectURL, err := url.Parse(pa.redirectURI)
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "invalid redirect_uri")
		return
	}
	q := redirectURL.Query()
	q.Set("code", proxyCode)
	if pa.state != "" {
		q.Set("state", pa.state)
	}
	redirectURL.RawQuery = q.Encode()

	// Clean up the pending flow.
	p.pending.Delete(flowID)

	respondJSON(w, http.StatusOK, map[string]any{
		"status":   "complete",
		"redirect": redirectURL.String(),
	})
}

func (p *oauthProxy) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid form body")
		return
	}

	switch r.FormValue("grant_type") {
	case "authorization_code":
		p.handleTokenAuthCode(w, r)
	case "refresh_token":
		p.handleTokenRefresh(w, r)
	default:
		respondOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "supported: authorization_code, refresh_token")
	}
}

func (p *oauthProxy) handleTokenAuthCode(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")
	clientID := r.FormValue("client_id")

	if code == "" || codeVerifier == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing code or code_verifier")
		return
	}

	// Look up and consume the proxy code (single-use).
	val, ok := p.codes.LoadAndDelete(code)
	if !ok {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown or expired code")
		return
	}

	entry := val.(*proxyCodeEntry)

	// Check expiry (5 minutes).
	if time.Since(entry.createdAt) > 5*time.Minute {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}

	if clientID != "" && clientID != entry.clientID {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}

	// Verify PKCE.
	if !verifyPKCE(codeVerifier, entry.codeChallenge, entry.codeChallengeMethod) {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":  entry.accessToken,
		"refresh_token": entry.refreshToken,
		"token_type":    "Bearer",
		"expires_in":    entry.expiresIn,
	})
}

func (p *oauthProxy) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing refresh_token")
		return
	}

	tc, err := htclient.TryRefresh(p.apiURL, refreshToken)
	if err != nil {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":  tc.AccessToken,
		"refresh_token": tc.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(tc.Expiry).Seconds()),
	})
}

// cleanup periodically removes stale entries from in-memory stores.
func (p *oauthProxy) cleanup() {
	for range time.Tick(5 * time.Minute) {
		cutoff := time.Now().Add(-10 * time.Minute)
		p.pending.Range(func(key, val any) bool {
			if val.(*pendingAuth).createdAt.Before(cutoff) {
				p.pending.Delete(key)
			}
			return true
		})
		p.codes.Range(func(key, val any) bool {
			if val.(*proxyCodeEntry).createdAt.Before(cutoff) {
				p.codes.Delete(key)
			}
			return true
		})
	}
}

// verifyPKCE checks the code_verifier against the stored code_challenge.
func verifyPKCE(verifier, challenge, method string) bool {
	if method != "S256" {
		return false
	}
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == challenge
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func respondOAuthError(w http.ResponseWriter, status int, errCode, desc string) {
	respondJSON(w, status, map[string]string{
		"error":             errCode,
		"error_description": desc,
	})
}

const authPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Hivetrack — Authenticate</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 480px; margin: 80px auto; text-align: center; color: #333; }
  a { color: #2563eb; font-weight: 600; }
  .status { margin-top: 24px; color: #666; }
  .spinner { display: inline-block; width: 16px; height: 16px; border: 2px solid #ccc; border-top-color: #2563eb; border-radius: 50%%; animation: spin 0.8s linear infinite; vertical-align: middle; margin-right: 8px; }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
</head>
<body>
  <h2>Authenticate with Hivetrack</h2>
  <p><a href="%s" id="auth-link" target="_blank">Click here to sign in</a></p>
  <p class="status"><span class="spinner"></span> <span id="msg">Waiting for authentication...</span></p>
  <script>
    window.open("%s", "_blank");
    const pollURL = "%s";
    (async function poll() {
      try {
        const res = await fetch(pollURL);
        const data = await res.json();
        if (data.status === "complete") {
          document.getElementById("msg").textContent = "Authenticated! Redirecting...";
          window.location.href = data.redirect;
          return;
        }
      } catch (e) {}
      setTimeout(poll, 2000);
    })();
  </script>
</body>
</html>`
