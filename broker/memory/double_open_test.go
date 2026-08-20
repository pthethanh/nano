package memory_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/pthethanh/nano/broker/memory"
)

func TestOpen_CalledTwiceDoesNotDoubleWorkerCount(t *testing.T) {
	br := memory.New[string](memory.Worker[string](50, 100))
	if err := br.Open(context.Background()); err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	defer br.Close(context.Background())

	settleGoroutineCount(t)
	before := runtime.NumGoroutine()

	if err := br.Open(context.Background()); err != nil {
		t.Fatalf("second Open() error = %v", err)
	}

	settleGoroutineCount(t)
	after := runtime.NumGoroutine()

	if after > before {
		t.Errorf("goroutine count grew from %d to %d after a second Open() call: it spawned another full set of worker goroutines instead of being a no-op", before, after)
	}
}

func settleGoroutineCount(t *testing.T) {
	t.Helper()
	runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
}
