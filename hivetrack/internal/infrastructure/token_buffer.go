package infrastructure

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

// TokenBuffer is a thread-safe in-memory store for partial streaming tokens
// emitted by Hivemind during refinement generation.
//
// Tokens are keyed by session UUID and accumulated until ClearPartialResponse
// is called (which happens when the final structured message arrives).
type TokenBuffer struct {
	mu      sync.Mutex
	tokens  map[uuid.UUID]*strings.Builder
	active  map[uuid.UUID]bool
}

func NewTokenBuffer() *TokenBuffer {
	return &TokenBuffer{
		tokens: make(map[uuid.UUID]*strings.Builder),
		active: make(map[uuid.UUID]bool),
	}
}

// Append appends a token chunk to the buffer for the given session.
func (b *TokenBuffer) Append(sessionID uuid.UUID, token string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.tokens[sessionID]; !ok {
		b.tokens[sessionID] = &strings.Builder{}
	}
	b.tokens[sessionID].WriteString(token)
	b.active[sessionID] = true
}

// Get returns the accumulated partial response for the session (empty if none).
func (b *TokenBuffer) Get(sessionID uuid.UUID) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if sb, ok := b.tokens[sessionID]; ok {
		return sb.String()
	}
	return ""
}

// IsGenerating reports whether tokens are currently being accumulated.
func (b *TokenBuffer) IsGenerating(sessionID uuid.UUID) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.active[sessionID]
}

// ClearPartialResponse clears the accumulated buffer and stops the generating state.
func (b *TokenBuffer) ClearPartialResponse(sessionID uuid.UUID) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.tokens, sessionID)
	delete(b.active, sessionID)
}
