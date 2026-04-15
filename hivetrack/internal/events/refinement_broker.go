package events

import (
	"sync"

	"github.com/google/uuid"
)

// RefinementBroker is an in-process fan-out hub that notifies subscribers
// when a refinement session belonging to a specific issue has changed.
// Events carry no payload — subscribers refetch the session through the
// normal query path on every notification.
type RefinementBroker struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan struct{}]struct{}
}

func NewRefinementBroker() *RefinementBroker {
	return &RefinementBroker{
		subs: make(map[uuid.UUID]map[chan struct{}]struct{}),
	}
}

// Subscribe returns a channel that receives a tick for every Publish against
// the given issue ID, and an unsubscribe function that must be called to
// release resources.
func (b *RefinementBroker) Subscribe(issueID uuid.UUID) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	b.mu.Lock()
	m, ok := b.subs[issueID]
	if !ok {
		m = make(map[chan struct{}]struct{})
		b.subs[issueID] = m
	}
	m[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() { b.unsubscribe(issueID, ch) }
}

func (b *RefinementBroker) unsubscribe(issueID uuid.UUID, ch chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	m, ok := b.subs[issueID]
	if !ok {
		return
	}
	if _, exists := m[ch]; !exists {
		return
	}
	delete(m, ch)
	close(ch)
	if len(m) == 0 {
		delete(b.subs, issueID)
	}
}

// Publish wakes every subscriber for the given issue ID. Subscribers whose
// channel already has a pending tick are left alone — a refetch coalesces
// multiple rapid events into one round-trip, and the polling fallback
// guarantees eventual consistency if a tick is dropped.
func (b *RefinementBroker) Publish(issueID uuid.UUID) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[issueID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
