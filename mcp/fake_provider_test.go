package mcp

import (
	"context"
	"fmt"
	"time"
)

// fakeProvider returns a preset result and records how many times it was called.
type fakeProvider struct {
	token tokenCache
	err   error
	calls int
}

func (f *fakeProvider) ProvideToken(_ context.Context) (tokenCache, error) {
	f.calls++
	return f.token, f.err
}

// randomTokenProvider returns a unique token on each call.
type randomTokenProvider struct {
	clock Clock
	calls int
}

func (r *randomTokenProvider) ProvideToken(_ context.Context) (tokenCache, error) {
	r.calls++
	return tokenCache{
		AccessToken: fmt.Sprintf("token-%d", r.calls),
		IssuedAt:    r.clock.Now(),
		Expiry:      r.clock.Now().Add(time.Hour),
	}, nil
}
