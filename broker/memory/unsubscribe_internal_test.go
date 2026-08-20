package memory

import (
	"context"
	"testing"

	"github.com/pthethanh/nano/broker"
)

// White-box test (same package) so it can inspect the internal subscriber
// map directly, since Unsubscribe()'s bug is that it doesn't touch it.
func TestUnsubscribe_RemovesSubscriberFromInternalMap(t *testing.T) {
	br := New[string]()
	if err := br.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer br.Close(context.Background())

	sub, err := br.Subscribe(context.Background(), "topic", func(broker.Event[string]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	br.mu.RLock()
	before := len(br.subs["topic"][""])
	br.mu.RUnlock()
	if before != 1 {
		t.Fatalf("expected 1 subscriber registered after Subscribe(), got %d", before)
	}

	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	br.mu.RLock()
	after := len(br.subs["topic"][""])
	br.mu.RUnlock()
	if after != 0 {
		t.Errorf("Unsubscribe() left %d subscriber(s) in the internal map, want 0 (subscriber map entry must be removed, not just flagged closed)", after)
	}
}
