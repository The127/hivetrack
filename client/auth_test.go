package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time       { return f.now }
func (f *fakeClock) Tick(d time.Duration) { f.now = f.now.Add(d) }

type fakeProvider struct {
	token TokenCache
	err   error
	calls int
}

func (f *fakeProvider) ProvideToken(_ context.Context) (TokenCache, error) {
	f.calls++
	return f.token, f.err
}

type randomTokenProvider struct {
	clock Clock
	calls int
}

func (r *randomTokenProvider) ProvideToken(_ context.Context) (TokenCache, error) {
	r.calls++
	return TokenCache{
		AccessToken: fmt.Sprintf("token-%d", r.calls),
		IssuedAt:    r.clock.Now(),
		Expiry:      r.clock.Now().Add(time.Hour),
	}, nil
}

func newTestCaching(inner TokenProvider, clock Clock, initial TokenCache, refreshFn func(string, string) (TokenCache, error)) *CachingTokenProvider {
	c := NewCachingTokenProvider(inner, clock, "http://example.com", initial, 0.1)
	if refreshFn != nil {
		c.RefreshFn = refreshFn
	}
	c.SaveFn = nil // don't write to disk in tests
	return c
}

func TestCachingTokenProvider_whenTokenIsFresh_returnsSameTokenOnRepeatedCalls(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	inner := &randomTokenProvider{clock: clock}

	c := newTestCaching(inner, clock, TokenCache{}, nil)

	first, err := c.ProvideToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := c.ProvideToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.AccessToken != second.AccessToken {
		t.Errorf("expected same token on second call, got %q then %q", first.AccessToken, second.AccessToken)
	}
	if inner.calls != 1 {
		t.Errorf("expected inner called once, got %d", inner.calls)
	}
}

func TestCachingTokenProvider_whenBelowRefreshThreshold_proactivelyRefreshes(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	initial := TokenCache{
		AccessToken:  "old-token",
		RefreshToken: "rt",
		IssuedAt:     clock.Now(),
		Expiry:       clock.Now().Add(100 * time.Minute),
	}
	clock.Tick(91 * time.Minute) // 9% remaining < 10% threshold
	inner := &fakeProvider{}

	refreshed := TokenCache{AccessToken: "refreshed-token", IssuedAt: clock.Now(), Expiry: clock.Now().Add(time.Hour)}
	c := newTestCaching(inner, clock, initial, func(_, _ string) (TokenCache, error) {
		return refreshed, nil
	})

	tc, err := c.ProvideToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.AccessToken != "refreshed-token" {
		t.Errorf("expected refreshed-token, got %s", tc.AccessToken)
	}
	if inner.calls != 0 {
		t.Error("inner should not be called when refresh succeeds")
	}
}

func TestCachingTokenProvider_whenRefreshFails_callsInner(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	initial := TokenCache{
		AccessToken:  "old-token",
		RefreshToken: "rt",
		IssuedAt:     clock.Now(),
		Expiry:       clock.Now().Add(100 * time.Minute),
	}
	clock.Tick(91 * time.Minute)
	innerToken := TokenCache{AccessToken: "inner-token", IssuedAt: clock.Now(), Expiry: clock.Now().Add(time.Hour)}
	inner := &fakeProvider{token: innerToken}

	c := newTestCaching(inner, clock, initial, func(_, _ string) (TokenCache, error) {
		return TokenCache{}, errors.New("refresh error")
	})

	tc, err := c.ProvideToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.AccessToken != "inner-token" {
		t.Errorf("expected inner-token, got %s", tc.AccessToken)
	}
	if inner.calls != 1 {
		t.Errorf("expected inner called once, got %d", inner.calls)
	}
}

func TestStaticTokenProvider(t *testing.T) {
	tc := TokenCache{AccessToken: "static"}
	p := &StaticTokenProvider{Token: tc}
	got, err := p.ProvideToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "static" {
		t.Errorf("expected static, got %s", got.AccessToken)
	}
}

func TestNewWithAuth(t *testing.T) {
	tc := TokenCache{AccessToken: "test-tok", Expiry: time.Now().Add(time.Hour)}
	c := NewWithAuth("http://example.com", &StaticTokenProvider{Token: tc})
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}
