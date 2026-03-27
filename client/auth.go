package client

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// TokenCache holds cached OIDC tokens.
type TokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IssuedAt     time.Time `json:"issued_at"`
	Expiry       time.Time `json:"expiry"`
	ServerURL    string    `json:"server_url"`
}

// Clock abstracts time.Now() for deterministic tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// RealClock is the production Clock implementation.
var RealClock Clock = realClock{}

// TokenProvider returns a valid TokenCache, blocking if necessary (e.g. device flow).
type TokenProvider interface {
	ProvideToken(ctx context.Context) (TokenCache, error)
}

// CachingTokenProvider wraps an inner TokenProvider. It returns the cached token
// while fresh (more than threshold percentage of lifetime remaining). When stale
// it attempts a token refresh first; only if that fails does it fall back to the
// inner provider (which may block on device flow).
type CachingTokenProvider struct {
	inner     TokenProvider
	clock     Clock
	RefreshFn func(baseURL, refreshToken string) (TokenCache, error)
	SaveFn    func(TokenCache) error
	baseURL   string
	threshold float64
	mu        sync.Mutex
	cached    TokenCache
}

// NewCachingTokenProvider constructs a CachingTokenProvider. threshold is the
// fraction of token lifetime below which a proactive refresh is attempted
// (e.g. 0.1 means refresh when less than 10% of lifetime remains).
func NewCachingTokenProvider(inner TokenProvider, clock Clock, baseURL string, initial TokenCache, threshold float64) *CachingTokenProvider {
	return &CachingTokenProvider{
		inner:     inner,
		clock:     clock,
		RefreshFn: TryRefresh,
		SaveFn:    SaveTokenFile,
		baseURL:   baseURL,
		threshold: threshold,
		cached:    initial,
	}
}

func (c *CachingTokenProvider) isFresh(tc TokenCache) bool {
	if tc.AccessToken == "" {
		return false
	}
	now := c.clock.Now()
	if !now.Before(tc.Expiry) {
		return false
	}
	if tc.IssuedAt.IsZero() {
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
func (c *CachingTokenProvider) ProvideToken(ctx context.Context) (TokenCache, error) {
	c.mu.Lock()
	if c.isFresh(c.cached) {
		tc := c.cached
		c.mu.Unlock()
		return tc, nil
	}
	rt := c.cached.RefreshToken
	c.mu.Unlock()

	if rt != "" {
		tc, err := c.RefreshFn(c.baseURL, rt)
		if err == nil {
			c.mu.Lock()
			c.cached = tc
			c.mu.Unlock()
			if c.SaveFn != nil {
				_ = c.SaveFn(tc)
			}
			return tc, nil
		}
		fmt.Fprintf(os.Stderr, "[hivetrack] token refresh failed: %v\n", err)
	}

	tc, err := c.inner.ProvideToken(ctx)
	if err == nil {
		c.mu.Lock()
		c.cached = tc
		c.mu.Unlock()
		if c.SaveFn != nil {
			_ = c.SaveFn(tc)
		}
	}
	return tc, err
}

// DeviceFlowProvider implements TokenProvider via OIDC device flow.
// The first call to ProvideToken initiates device flow and returns an error
// containing the auth URL. Once the user authenticates, subsequent calls
// return the token.
type DeviceFlowProvider struct {
	BaseURL string

	mu      sync.Mutex
	flow    *DeviceFlow
	result  *TokenCache
	flowErr error
	expires time.Time
}

// ProvideToken starts device flow on first call and returns an error with the auth URL.
// Once the user completes auth, the next call returns the cached token.
func (f *DeviceFlowProvider) ProvideToken(_ context.Context) (TokenCache, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.result != nil && time.Now().Before(f.expires) {
		tc := *f.result
		f.result = nil
		f.flow = nil
		return tc, nil
	}

	if f.flowErr != nil || (f.result != nil && !time.Now().Before(f.expires)) {
		f.flow = nil
		f.result = nil
		f.flowErr = nil
	}

	if f.flow == nil {
		flow, err := InitDeviceFlow(f.BaseURL)
		if err != nil {
			return TokenCache{}, fmt.Errorf("failed to start device flow: %w", err)
		}
		f.flow = flow
		go func() {
			tc, err := flow.WaitForToken(context.Background())
			f.mu.Lock()
			defer f.mu.Unlock()
			if err != nil {
				f.flowErr = err
				f.flow = nil
			} else {
				f.result = &tc
				f.expires = tc.Expiry
			}
		}()
	}

	authURL := f.flow.VerificationURIComplete
	if authURL == "" {
		authURL = f.flow.VerificationURI
	}
	return TokenCache{}, fmt.Errorf("not authenticated: open %s to authenticate with Hivetrack, then retry", authURL)
}

// StaticTokenProvider returns a fixed token. Useful for testing and CI.
type StaticTokenProvider struct {
	Token TokenCache
}

func (s *StaticTokenProvider) ProvideToken(_ context.Context) (TokenCache, error) {
	return s.Token, nil
}

// NewWithAuth creates a client using a TokenProvider for authentication.
// This is the recommended way to create an authenticated client.
func NewWithAuth(baseURL string, provider TokenProvider) *Client {
	return New(baseURL, func(ctx context.Context) (string, error) {
		tc, err := provider.ProvideToken(ctx)
		if err != nil {
			return "", err
		}
		return tc.AccessToken, nil
	})
}
