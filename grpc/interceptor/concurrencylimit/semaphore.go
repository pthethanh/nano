package concurrencylimit

import (
	"context"

	"google.golang.org/grpc/status"
)

// Semaphore is a bounded in-memory limiter.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore returns a limiter with the provided maximum concurrency.
func NewSemaphore(limit int) *Semaphore {
	if limit <= 0 {
		limit = 1
	}
	return &Semaphore{ch: make(chan struct{}, limit)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		// Map to a proper gRPC status code (Canceled/DeadlineExceeded)
		// instead of surfacing the raw context error as codes.Unknown.
		return status.FromContextError(ctx.Err()).Err()
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}
