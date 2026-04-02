package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/server"

	htclient "github.com/the127/hivetrack/client"
)

// sessionAuthManager implements TokenProvider with per-MCP-session device flow.
// Each session gets its own OIDC device flow and cached token.
// If the caller provides a Bearer token via context, it is used directly.
type sessionAuthManager struct {
	apiURL   string
	sessions sync.Map // map[string]*sessionState
}

func newSessionAuthManager(apiURL string) *sessionAuthManager {
	return &sessionAuthManager{apiURL: apiURL}
}

// sessionState tracks auth for a single MCP session.
type sessionState struct {
	mu sync.Mutex

	// Device flow (non-nil while waiting for user to authenticate).
	flow    *htclient.DeviceFlow
	authURL string

	// Background polling cancellation.
	pollCancel context.CancelFunc

	// Cached token (non-nil once authenticated).
	token *htclient.TokenCache
}

// ProvideToken returns a token for the current request.
// Priority: Bearer from HTTP header > cached session token > device flow.
func (m *sessionAuthManager) ProvideToken(ctx context.Context) (htclient.TokenCache, error) {
	// 1. Bearer passthrough: if caller sent an Authorization header, use it.
	if token, ok := ctx.Value(bearerTokenKey{}).(string); ok && token != "" {
		return htclient.TokenCache{AccessToken: token}, nil
	}

	// 2. Session-based device flow.
	sess := server.ClientSessionFromContext(ctx)
	if sess == nil {
		return htclient.TokenCache{}, fmt.Errorf("not authenticated: no MCP session and no Bearer token")
	}

	st := m.getOrCreateSession(sess.SessionID())
	return st.provideToken(m.apiURL)
}

func (m *sessionAuthManager) getOrCreateSession(id string) *sessionState {
	if v, ok := m.sessions.Load(id); ok {
		return v.(*sessionState)
	}
	st := &sessionState{}
	if v, loaded := m.sessions.LoadOrStore(id, st); loaded {
		return v.(*sessionState)
	}
	return st
}

// removeSession cancels any in-flight polling and removes session state.
func (m *sessionAuthManager) removeSession(id string) {
	if v, ok := m.sessions.LoadAndDelete(id); ok {
		st := v.(*sessionState)
		st.mu.Lock()
		if st.pollCancel != nil {
			st.pollCancel()
		}
		st.mu.Unlock()
	}
}

// provideToken returns a token for a session, initiating device flow if needed.
func (st *sessionState) provideToken(apiURL string) (htclient.TokenCache, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Have a valid token? Return it.
	if st.token != nil && time.Now().Before(st.token.Expiry) {
		return *st.token, nil
	}

	// Token expired but we have a refresh token? Try refresh.
	if st.token != nil && st.token.RefreshToken != "" {
		tc, err := htclient.TryRefresh(apiURL, st.token.RefreshToken)
		if err == nil {
			st.token = &tc
			return tc, nil
		}
		// Refresh failed — clear stale token and fall through to device flow.
		st.token = nil
		fmt.Printf("[mcp] token refresh failed: %v\n", err)
	}

	// Device flow already in progress? Tell caller to authenticate.
	if st.flow != nil {
		return htclient.TokenCache{}, fmt.Errorf(
			"not authenticated: open %s to authenticate with Hivetrack, then retry",
			st.authURL,
		)
	}

	// Start new device flow.
	flow, err := htclient.InitDeviceFlow(apiURL)
	if err != nil {
		return htclient.TokenCache{}, fmt.Errorf("starting device flow: %w", err)
	}

	authURL := flow.VerificationURIComplete
	if authURL == "" {
		authURL = flow.VerificationURI
	}

	st.flow = flow
	st.authURL = authURL

	// Poll in background so the token is ready when the user retries.
	ctx, cancel := context.WithCancel(context.Background())
	st.pollCancel = cancel
	go st.pollForToken(ctx, flow)

	return htclient.TokenCache{}, fmt.Errorf(
		"not authenticated: open %s to authenticate with Hivetrack, then retry",
		authURL,
	)
}

// pollForToken runs in a goroutine, polling the OIDC token endpoint.
func (st *sessionState) pollForToken(ctx context.Context, flow *htclient.DeviceFlow) {
	tc, err := flow.WaitForToken(ctx)

	st.mu.Lock()
	defer st.mu.Unlock()

	if err != nil {
		// Flow failed or was cancelled — reset so next call starts fresh.
		st.flow = nil
		st.pollCancel = nil
		return
	}

	st.token = &tc
	st.flow = nil
	st.pollCancel = nil
}
