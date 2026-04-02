package main

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	htclient "github.com/the127/hivetrack/client"
)

// fakeSession implements server.ClientSession for testing.
type fakeSession struct {
	id string
	ch chan mcp.JSONRPCNotification
}

func (f fakeSession) Initialize()                                       {}
func (f fakeSession) Initialized() bool                                 { return true }
func (f fakeSession) NotificationChannel() chan<- mcp.JSONRPCNotification { return f.ch }
func (f fakeSession) SessionID() string                                 { return f.id }

func newFakeSession(id string) fakeSession {
	return fakeSession{id: id, ch: make(chan mcp.JSONRPCNotification, 1)}
}

func ctxWithSession(sess server.ClientSession) context.Context {
	s := server.NewMCPServer("test", "1.0.0")
	return s.WithContext(context.Background(), sess)
}

func TestBearerPassthrough(t *testing.T) {
	mgr := newSessionAuthManager("http://localhost:8086")

	ctx := context.WithValue(context.Background(), bearerTokenKey{}, "my-token")
	tc, err := mgr.ProvideToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.AccessToken != "my-token" {
		t.Fatalf("expected my-token, got %s", tc.AccessToken)
	}
}

func TestNoSessionNoBearer(t *testing.T) {
	mgr := newSessionAuthManager("http://localhost:8086")

	_, err := mgr.ProvideToken(context.Background())
	if err == nil {
		t.Fatal("expected error for missing session and bearer")
	}
}

func TestSessionReturnsExistingToken(t *testing.T) {
	mgr := newSessionAuthManager("http://localhost:8086")

	// Pre-populate a session with a valid token.
	sess := newFakeSession("test-session")
	st := mgr.getOrCreateSession(sess.id)
	future := htclient.TokenCache{
		AccessToken: "cached-token",
		Expiry:      htclient.RealClock.Now().Add(time.Hour),
	}
	st.token = &future

	ctx := ctxWithSession(sess)
	tc, err := mgr.ProvideToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.AccessToken != "cached-token" {
		t.Fatalf("expected cached-token, got %s", tc.AccessToken)
	}
}

func TestSessionIsolation(t *testing.T) {
	mgr := newSessionAuthManager("http://localhost:8086")

	// Two different sessions should have independent state.
	s1 := mgr.getOrCreateSession("session-1")
	s2 := mgr.getOrCreateSession("session-2")

	tok := htclient.TokenCache{
		AccessToken: "s1-token",
		Expiry:      htclient.RealClock.Now().Add(time.Hour),
	}
	s1.token = &tok

	if s2.token != nil {
		t.Fatal("session-2 should not have a token")
	}
}

func TestRemoveSessionCleansUp(t *testing.T) {
	mgr := newSessionAuthManager("http://localhost:8086")

	mgr.getOrCreateSession("doomed")
	mgr.removeSession("doomed")

	if _, ok := mgr.sessions.Load("doomed"); ok {
		t.Fatal("session should have been removed")
	}
}
