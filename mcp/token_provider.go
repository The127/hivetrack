package mcp

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// TokenProvider returns a valid tokenCache, blocking if necessary (e.g. device flow).
type TokenProvider interface {
	ProvideToken(ctx context.Context) (tokenCache, error)
}

// CachingTokenProvider wraps an inner TokenProvider. It returns the cached token
// while fresh (more than threshold percentage of lifetime remaining). When stale
// it attempts a token refresh first; only if that fails does it fall back to the
// inner provider (which may block on device flow).
type CachingTokenProvider struct {
	inner     TokenProvider
	clock     Clock
	refreshFn func(baseURL, refreshToken string) (tokenCache, error)
	baseURL   string
	threshold float64 // refresh when remaining lifetime fraction drops below this (e.g. 0.1 = 10%)
	mu        sync.Mutex
	cached    tokenCache
}

// NewCachingTokenProvider constructs a CachingTokenProvider. threshold is the
// fraction of token lifetime below which a proactive refresh is attempted
// (e.g. 0.1 means refresh when less than 10% of lifetime remains).
func NewCachingTokenProvider(inner TokenProvider, clock Clock, baseURL string, initial tokenCache, threshold float64) *CachingTokenProvider {
	return &CachingTokenProvider{
		inner:     inner,
		clock:     clock,
		refreshFn: tryRefresh,
		baseURL:   baseURL,
		threshold: threshold,
		cached:    initial,
	}
}

func (c *CachingTokenProvider) isFresh(tc tokenCache) bool {
	if tc.AccessToken == "" {
		return false
	}
	now := c.clock.Now()
	if !now.Before(tc.Expiry) {
		return false
	}
	if tc.IssuedAt.IsZero() {
		// No issuance time recorded (e.g. old cache file); fall back to expiry-only check.
		return true
	}
	lifetime := tc.Expiry.Sub(tc.IssuedAt)
	if lifetime <= 0 {
		return false
	}
	remaining := tc.Expiry.Sub(now)
	return float64(remaining)/float64(lifetime) > c.threshold
}

// ProvideToken returns the cached token if still fresh; otherwise tries to refresh,
// falling back to the inner provider if refresh fails or no refresh token exists.
func (c *CachingTokenProvider) ProvideToken(ctx context.Context) (tokenCache, error) {
	c.mu.Lock()
	if c.isFresh(c.cached) {
		tc := c.cached
		c.mu.Unlock()
		return tc, nil
	}
	rt := c.cached.RefreshToken
	c.mu.Unlock()

	if rt != "" {
		tc, err := c.refreshFn(c.baseURL, rt)
		if err == nil {
			c.mu.Lock()
			c.cached = tc
			c.mu.Unlock()
			_ = saveCache(tc)
			return tc, nil
		}
		fmt.Fprintf(os.Stderr, "[mcp] token refresh failed: %v\n", err)
	}

	tc, err := c.inner.ProvideToken(ctx)
	if err == nil {
		c.mu.Lock()
		c.cached = tc
		c.mu.Unlock()
		_ = saveCache(tc)
	}
	return tc, err
}

// DeviceFlowProvider implements TokenProvider via OIDC device flow.
// It logs the auth URL to stderr and blocks until the user completes auth.
type DeviceFlowProvider struct {
	BaseURL string
}

// ProvideToken initiates device flow, prints the auth URL, and waits for completion.
func (f *DeviceFlowProvider) ProvideToken(ctx context.Context) (tokenCache, error) {
	flow, err := InitDeviceFlow(f.BaseURL)
	if err != nil {
		return tokenCache{}, err
	}
	authURL := flow.VerificationURIComplete
	if authURL == "" {
		authURL = flow.VerificationURI
	}
	fmt.Fprintf(os.Stderr, "[mcp] authenticate at: %s\n", authURL)
	return flow.WaitForToken(ctx)
}
