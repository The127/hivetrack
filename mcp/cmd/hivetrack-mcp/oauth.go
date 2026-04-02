package main

import (
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

// oauthProxy implements an OAuth 2.1 proxy that forwards to the real OIDC provider.
// MCP clients discover this via /.well-known/oauth-authorization-server,
// register dynamically, and go through an authorization code + PKCE flow.
// Behind the scenes, the proxy redirects to the real OIDC provider's authorize
// endpoint and exchanges codes on behalf of the client.
type oauthProxy struct {
	apiURL      string // Hivetrack API URL (for OIDC discovery)
	externalURL string // This server's public URL

	// OIDC provider endpoints (cached).
	oidcMu            sync.RWMutex
	authorizeEndpoint string
	tokenEndpoint     string
	clientID          string
	lastDiscovery     time.Time

	pending sync.Map // proxyState (string) -> *pendingAuth
	codes   sync.Map // proxyCode (string) -> *proxyCodeEntry
}

type pendingAuth struct {
	clientRedirectURI   string
	clientState         string
	codeChallenge       string
	codeChallengeMethod string
	createdAt           time.Time
}

type proxyCodeEntry struct {
	accessToken  string
	refreshToken string
	expiresIn    int

	codeChallenge       string
	codeChallengeMethod string
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
	mux.HandleFunc("GET /oauth/callback", p.handleCallback)
	mux.HandleFunc("POST /oauth/token", p.handleToken)
}

// discoverOIDC fetches and caches the OIDC provider's endpoints.
func (p *oauthProxy) discoverOIDC() (authorizeEndpoint, tokenEndpoint, clientID string, err error) {
	p.oidcMu.RLock()
	if time.Since(p.lastDiscovery) < 5*time.Minute && p.authorizeEndpoint != "" {
		defer p.oidcMu.RUnlock()
		return p.authorizeEndpoint, p.tokenEndpoint, p.clientID, nil
	}
	p.oidcMu.RUnlock()

	providerCfg, err := htclient.FetchOIDCProviderConfig(p.apiURL)
	if err != nil {
		return "", "", "", fmt.Errorf("fetching OIDC config: %w", err)
	}

	doc, err := htclient.FetchOIDCDiscovery(providerCfg.Authority)
	if err != nil {
		return "", "", "", fmt.Errorf("fetching OIDC discovery: %w", err)
	}

	authzEP, _ := doc["authorization_endpoint"].(string)
	tokenEP, _ := doc["token_endpoint"].(string)
	if authzEP == "" || tokenEP == "" {
		return "", "", "", fmt.Errorf("OIDC provider missing authorization_endpoint or token_endpoint")
	}

	p.oidcMu.Lock()
	p.authorizeEndpoint = authzEP
	p.tokenEndpoint = tokenEP
	p.clientID = providerCfg.ClientID
	p.lastDiscovery = time.Now()
	p.oidcMu.Unlock()

	return authzEP, tokenEP, providerCfg.ClientID, nil
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

	// Discover the real client_id from the OIDC provider.
	_, _, clientID, err := p.discoverOIDC()
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "OIDC discovery failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"client_id":                  clientID,
		"client_id_issued_at":        time.Now().Unix(),
		"redirect_uris":              body.RedirectURIs,
		"client_name":                body.ClientName,
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}

func (p *oauthProxy) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	if redirectURI == "" || codeChallenge == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing required parameters")
		return
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	authorizeEndpoint, _, clientID, err := p.discoverOIDC()
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "OIDC discovery failed: "+err.Error())
		return
	}

	// Generate a proxy state to link the callback back to this request.
	proxyState := randomHex(16)
	p.pending.Store(proxyState, &pendingAuth{
		clientRedirectURI:   redirectURI,
		clientState:         state,
		codeChallenge:       codeChallenge,
		codeChallengeMethod: codeChallengeMethod,
		createdAt:           time.Now(),
	})

	// Redirect to the real OIDC provider's authorize endpoint.
	u, _ := url.Parse(authorizeEndpoint)
	params := u.Query()
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", p.externalURL+"/oauth/callback")
	params.Set("state", proxyState)
	params.Set("scope", "openid offline_access")
	u.RawQuery = params.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (p *oauthProxy) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	proxyState := q.Get("state")

	if errCode := q.Get("error"); errCode != "" {
		desc := q.Get("error_description")
		if desc == "" {
			desc = "authorization failed"
		}
		respondOAuthError(w, http.StatusBadRequest, errCode, desc)
		return
	}

	if code == "" || proxyState == "" {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "missing code or state")
		return
	}

	val, ok := p.pending.LoadAndDelete(proxyState)
	if !ok {
		respondOAuthError(w, http.StatusBadRequest, "invalid_request", "unknown or expired state")
		return
	}
	pa := val.(*pendingAuth)

	// Exchange the real auth code at the OIDC provider's token endpoint.
	_, tokenEndpoint, clientID, err := p.discoverOIDC()
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "OIDC discovery failed: "+err.Error())
		return
	}

	resp, err := http.PostForm(tokenEndpoint, url.Values{ //nolint:noctx
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {p.externalURL + "/oauth/callback"},
		"client_id":    {clientID},
	})
	if err != nil {
		respondOAuthError(w, http.StatusBadGateway, "server_error", "token exchange failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		respondOAuthError(w, http.StatusBadGateway, "server_error", "invalid token response: "+err.Error())
		return
	}
	if tr.Error != "" {
		respondOAuthError(w, http.StatusBadGateway, "server_error", "token exchange error: "+tr.Error+": "+tr.ErrorDesc)
		return
	}

	// Store tokens under a proxy code for the MCP client to exchange.
	proxyCode := randomHex(24)
	p.codes.Store(proxyCode, &proxyCodeEntry{
		accessToken:         tr.AccessToken,
		refreshToken:        tr.RefreshToken,
		expiresIn:           tr.ExpiresIn,
		codeChallenge:       pa.codeChallenge,
		codeChallengeMethod: pa.codeChallengeMethod,
		createdAt:           time.Now(),
	})

	// Redirect back to the MCP client with the proxy code.
	redirectURL, err := url.Parse(pa.clientRedirectURI)
	if err != nil {
		respondOAuthError(w, http.StatusInternalServerError, "server_error", "invalid client redirect_uri")
		return
	}
	rq := redirectURL.Query()
	rq.Set("code", proxyCode)
	if pa.clientState != "" {
		rq.Set("state", pa.clientState)
	}
	redirectURL.RawQuery = rq.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
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

	if time.Since(entry.createdAt) > 5*time.Minute {
		respondOAuthError(w, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}

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
