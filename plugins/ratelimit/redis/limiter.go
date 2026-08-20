// Package redis provides a distributed, Redis-backed rate limiter
// implementing github.com/pthethanh/nano/grpc/interceptor/ratelimit.Limiter.
//
// Unlike ratelimit.NewTokenBucket (in-process only, per-replica), this
// limiter's state lives in Redis and is shared across every replica of a
// horizontally-scaled service, so the configured limit is enforced across
// the whole fleet rather than per instance.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/ratelimit"
	goredis "github.com/redis/go-redis/v9"
)

// Limiter is a fixed-window rate limiter: up to Limit calls are allowed per
// key within each Window, using a Redis INCR+EXPIRE counter. It's simpler
// and cheaper than a sliding window or token bucket, at the cost of
// allowing up to 2x Limit calls across a window boundary (e.g. a burst
// right at the end of one window followed by another right at the start of
// the next). For most API rate-limiting use cases that trade-off is fine;
// use a different Limiter implementation if you need exact enforcement.
type Limiter struct {
	client  goredis.UniversalClient
	opts    *goredis.UniversalOptions
	managed bool
	keyFunc func(ctx context.Context) string
	prefix  string
	limit   int64
	window  time.Duration
}

var _ ratelimit.Limiter = (*Limiter)(nil)

// ErrNotOpen is returned by Allow if Open has not been called (or failed).
var ErrNotOpen = errors.New("ratelimit/redis: not open")

// New returns a Limiter. Defaults: 127.0.0.1:6379, 100 calls per minute
// across a single global key (see KeyFunc to key by subject/IP/method).
func New(opts ...Option) *Limiter {
	l := &Limiter{
		opts: &goredis.UniversalOptions{
			Addrs: []string{"127.0.0.1:6379"},
		},
		managed: true,
		keyFunc: func(ctx context.Context) string { return "global" },
		prefix:  "ratelimit",
		limit:   100,
		window:  time.Minute,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Open establishes the connection to Redis.
func (l *Limiter) Open(ctx context.Context) error {
	if l.client != nil {
		return l.client.Ping(ctx).Err()
	}
	l.client = goredis.NewUniversalClient(l.opts)
	if err := l.client.Ping(ctx).Err(); err != nil {
		_ = l.client.Close()
		l.client = nil
		return err
	}
	l.managed = true
	return nil
}

// Close closes the managed Redis client. It is a no-op for a client
// injected via the Client option.
func (l *Limiter) Close(ctx context.Context) error {
	if l.client == nil {
		return nil
	}
	if !l.managed {
		l.client = nil
		return nil
	}
	err := l.client.Close()
	l.client = nil
	return err
}

// Allow implements ratelimit.Limiter. It returns ratelimit.ErrLimited
// (codes.ResourceExhausted) once the configured limit is exceeded within
// the current window for the key derived from ctx by KeyFunc.
func (l *Limiter) Allow(ctx context.Context) error {
	if l.client == nil {
		return ErrNotOpen
	}
	key := l.prefix + ":" + l.keyFunc(ctx)
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("ratelimit/redis: %w", err)
	}
	if count == 1 {
		if err := l.client.Expire(ctx, key, l.window).Err(); err != nil {
			return fmt.Errorf("ratelimit/redis: %w", err)
		}
	}
	if count > l.limit {
		return ratelimit.ErrLimited
	}
	return nil
}
