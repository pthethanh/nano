package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Option configures a Limiter.
type Option func(*Limiter)

// Address configures Redis server addresses.
func Address(addrs ...string) Option {
	return func(l *Limiter) {
		l.opts.Addrs = append([]string(nil), addrs...)
	}
}

// Limit sets the maximum number of Allow calls permitted per key within
// window. Values <= 0 are ignored (the default/previous value is kept).
func Limit(n int64, window time.Duration) Option {
	return func(l *Limiter) {
		if n > 0 {
			l.limit = n
		}
		if window > 0 {
			l.window = window
		}
	}
}

// KeyFunc configures how a request is mapped to a rate-limit bucket key,
// e.g. per-subject, per-client-IP, or per-method. The default groups every
// call into a single global bucket.
func KeyFunc(fn func(ctx context.Context) string) Option {
	return func(l *Limiter) {
		if fn != nil {
			l.keyFunc = fn
		}
	}
}

// Prefix sets the Redis key prefix used for all rate-limit keys (default "ratelimit").
func Prefix(prefix string) Option {
	return func(l *Limiter) {
		l.prefix = prefix
	}
}

// Client injects an existing Redis client instead of letting Open construct one.
func Client(client goredis.UniversalClient) Option {
	return func(l *Limiter) {
		l.client = client
		l.managed = false
	}
}

// Options replaces the Redis universal options.
func Options(opts *goredis.UniversalOptions) Option {
	return func(l *Limiter) {
		if opts == nil {
			return
		}
		clone := *opts
		if opts.Addrs != nil {
			clone.Addrs = append([]string(nil), opts.Addrs...)
		}
		l.opts = &clone
	}
}
