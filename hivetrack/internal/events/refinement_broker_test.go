package events

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRefinementBroker_PublishDeliversToSubscriber(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	ch, unsub := b.Subscribe(issueID)
	defer unsub()

	b.Publish(issueID)

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected tick after Publish")
	}
}

func TestRefinementBroker_PublishToUnrelatedIssueIsIgnored(t *testing.T) {
	b := NewRefinementBroker()
	issueA, issueB := uuid.New(), uuid.New()

	ch, unsub := b.Subscribe(issueA)
	defer unsub()

	b.Publish(issueB)

	select {
	case <-ch:
		t.Fatal("subscriber to issue A must not receive tick for issue B")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRefinementBroker_MultipleSubscribersReceiveTick(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	ch1, unsub1 := b.Subscribe(issueID)
	ch2, unsub2 := b.Subscribe(issueID)
	defer unsub1()
	defer unsub2()

	b.Publish(issueID)

	for i, ch := range []<-chan struct{}{ch1, ch2} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d did not receive tick", i)
		}
	}
}

func TestRefinementBroker_UnsubscribeStopsDelivery(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	ch, unsub := b.Subscribe(issueID)
	unsub()

	// Channel is closed after unsubscribe — reading returns zero value immediately.
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel should be closed after unsubscribe")
		}
	case <-time.After(time.Second):
		t.Fatal("expected channel to be closed")
	}

	// Publishing after unsubscribe must not panic.
	b.Publish(issueID)
}

func TestRefinementBroker_DoubleUnsubscribeIsSafe(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	_, unsub := b.Subscribe(issueID)
	unsub()
	unsub() // second call must be a no-op
}

func TestRefinementBroker_SlowConsumerIsNotBlocked(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	_, unsub := b.Subscribe(issueID)
	defer unsub()

	// Publish many times without reading — buffered channel holds one,
	// subsequent publishes are dropped, nothing blocks.
	done := make(chan struct{})
	go func() {
		for range 1000 {
			b.Publish(issueID)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on slow consumer")
	}
}

func TestRefinementBroker_ConcurrentSubscribeAndPublish(t *testing.T) {
	b := NewRefinementBroker()
	issueID := uuid.New()

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			ch, unsub := b.Subscribe(issueID)
			defer unsub()
			b.Publish(issueID)
			select {
			case <-ch:
			case <-time.After(time.Second):
			}
		})
	}
	wg.Wait()
}
