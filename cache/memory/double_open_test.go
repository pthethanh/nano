package memory_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/pthethanh/nano/cache/memory"
)

func TestOpen_CalledTwiceDoesNotStartASecondSweepGoroutine(t *testing.T) {
	c := memory.New[string, []byte]()
	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	defer c.Close(context.Background())

	settleGoroutineCount(t)
	before := runtime.NumGoroutine()

	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("second Open() error = %v", err)
	}

	settleGoroutineCount(t)
	after := runtime.NumGoroutine()

	if after > before {
		t.Errorf("goroutine count grew from %d to %d after a second Open() call: it started another background sweep goroutine instead of being a no-op", before, after)
	}
}

func settleGoroutineCount(t *testing.T) {
	t.Helper()
	runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
}
