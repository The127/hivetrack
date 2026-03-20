package mcp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newCachingFetcher(inner TokenFetcher, clock Clock, initial tokenCache, refreshFn func(string, string) (tokenCache, error)) *CachingTokenFetcher {
	c := NewCachingTokenFetcher(inner, clock, "http://example.com", initial, 0.1)
	if refreshFn != nil {
		c.refreshFn = refreshFn
	}
	return c
}

func TestCachingTokenFetcher_whenTokenIsFresh_returnsSameTokenOnRepeatedCalls(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	inner := &randomTokenFetcher{clock: clock}

	c := newCachingFetcher(inner, clock, tokenCache{}, nil)

	first, err := c.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := c.FetchToken(context.Background())
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

func TestCachingTokenFetcher_whenBelowRefreshThreshold_proactivelyRefreshes(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	// 5% lifetime remaining → below 10% threshold → refresh expected
	initial := tokenCache{
		AccessToken:  "old-token",
		RefreshToken: "rt",
		IssuedAt:     clock.now.Add(-95 * time.Minute),
		Expiry:       clock.now.Add(5 * time.Minute),
	}
	inner := &fakeFetcher{}

	refreshed := tokenCache{AccessToken: "refreshed-token", IssuedAt: clock.now, Expiry: clock.now.Add(time.Hour)}
	c := newCachingFetcher(inner, clock, initial, func(_, _ string) (tokenCache, error) {
		return refreshed, nil
	})

	tc, err := c.FetchToken(context.Background())
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

func TestCachingTokenFetcher_whenRefreshFails_callsInner(t *testing.T) {
	clock := &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	initial := tokenCache{
		AccessToken:  "old-token",
		RefreshToken: "rt",
		IssuedAt:     clock.now.Add(-95 * time.Minute),
		Expiry:       clock.now.Add(5 * time.Minute),
	}
	innerToken := tokenCache{AccessToken: "inner-token", IssuedAt: clock.now, Expiry: clock.now.Add(time.Hour)}
	inner := &fakeFetcher{token: innerToken}

	c := newCachingFetcher(inner, clock, initial, func(_, _ string) (tokenCache, error) {
		return tokenCache{}, errors.New("refresh error")
	})

	tc, err := c.FetchToken(context.Background())
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
