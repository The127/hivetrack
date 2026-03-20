package mcp

import (
	"context"
	"fmt"
	"time"
)

// fakeFetcher returns a preset result and records how many times it was called.
type fakeFetcher struct {
	token tokenCache
	err   error
	calls int
}

func (f *fakeFetcher) FetchToken(_ context.Context) (tokenCache, error) {
	f.calls++
	return f.token, f.err
}

// randomTokenFetcher returns a unique token on each call.
type randomTokenFetcher struct {
	clock Clock
	calls int
}

func (r *randomTokenFetcher) FetchToken(_ context.Context) (tokenCache, error) {
	r.calls++
	return tokenCache{
		AccessToken: fmt.Sprintf("token-%d", r.calls),
		IssuedAt:    r.clock.Now(),
		Expiry:      r.clock.Now().Add(time.Hour),
	}, nil
}
