package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
)

// TokenBucket is an in-memory token-bucket rate limiter.
//
// It rejects requests immediately when no token is available.
type TokenBucket struct {
	limiter *rate.Limiter
}

// NewTokenBucket returns a limiter that refills tokens at the provided rate
// and allows bursts up to burst.
func NewTokenBucket(limit rate.Limit, burst int) *TokenBucket {
	if burst <= 0 {
		burst = 1
	}
	return &TokenBucket{
		limiter: rate.NewLimiter(limit, burst),
	}
}

// Allow reports whether the current request can proceed immediately.
func (t *TokenBucket) Allow(context.Context) error {
	if t == nil || t.limiter == nil {
		return ErrLimited
	}
	if !t.limiter.Allow() {
		return ErrLimited
	}
	return nil
}
