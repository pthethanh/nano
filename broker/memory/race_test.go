package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/broker/memory"
)

// This test only proves the bug when run with `go test -race`: concurrent
// Open/Close/Publish/CheckHealth calls read and write br.opened without any
// synchronization.
func TestConcurrentOpenCloseCheckHealthDoesNotRace(t *testing.T) {
	br := memory.New[string](memory.Worker[string](4, 16))

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = br.Open(context.Background())
		}()
	}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = br.CheckHealth(context.Background())
		}()
	}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = br.Publish(context.Background(), "topic", ptr("msg"))
		}()
	}
	wg.Wait()
	_ = br.Close(context.Background())
}

func ptr[T any](v T) *T { return &v }

// Worker(0, n) must not silently produce a broker with zero consumer
// goroutines: doing so makes every Publish() block forever once the channel
// buffer fills (immediately, if buffer is also 0).
func TestWorker_ZeroWorkersDoesNotDeadlockPublish(t *testing.T) {
	br := memory.New[string](memory.Worker[string](0, 0))
	if err := br.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer br.Close(context.Background())

	// A subscriber is required so Publish() actually attempts to hand the
	// message off through br.ch instead of returning early.
	if _, err := br.Subscribe(context.Background(), "topic", func(ev broker.Event[string]) error { return nil }); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- br.Publish(context.Background(), "topic", ptr("msg"))
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Publish() blocked forever: Worker(0, 0) produced a broker with no consumer goroutines")
	}
}
