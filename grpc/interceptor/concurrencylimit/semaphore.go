package concurrencylimit

import "context"

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
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}
