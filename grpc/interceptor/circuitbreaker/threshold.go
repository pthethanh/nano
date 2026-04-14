package circuitbreaker

import (
	"sync"
	"time"
)

// ThresholdBreaker opens after N consecutive failures and closes after the reset timeout.
type ThresholdBreaker struct {
	mu                  sync.Mutex
	consecutiveFailures int
	maxFailures         int
	resetTimeout        time.Duration
	openedAt            time.Time
	shouldTrip          func(error) bool
}

// NewThresholdBreaker returns a simple consecutive-failure circuit breaker.
func NewThresholdBreaker(maxFailures int, resetTimeout time.Duration, opts ...Option) *ThresholdBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	o := &options{shouldTrip: defaultShouldTrip}
	for _, opt := range opts {
		opt(o)
	}
	return &ThresholdBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		shouldTrip:   o.shouldTrip,
	}
}

func (b *ThresholdBreaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consecutiveFailures < b.maxFailures {
		return nil
	}
	if time.Since(b.openedAt) >= b.resetTimeout {
		b.consecutiveFailures = 0
		b.openedAt = time.Time{}
		return nil
	}
	return ErrOpen
}

func (b *ThresholdBreaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveFailures = 0
	b.openedAt = time.Time{}
}

func (b *ThresholdBreaker) Failure(err error) {
	if b == nil || b.shouldTrip == nil || !b.shouldTrip(err) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveFailures++
	if b.consecutiveFailures >= b.maxFailures && b.openedAt.IsZero() {
		b.openedAt = time.Now()
	}
}
