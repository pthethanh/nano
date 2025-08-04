package cache

import (
	"context"
	"time"
)

type (
	// SetOptions configures options for setting a cache value.
	SetOptions struct {
		TTL time.Duration // Time To Live for the key.
	}

	// SetOption modifies SetOptions.
	SetOption func(*SetOptions)

	// Cacher defines a generic cache service.
	Cacher[K comparable, V any] interface {
		// Open connects to the cache backend.
		Open(ctx context.Context) error
		// Get retrieves a value. Returns ErrNotFound if key is missing.
		Get(ctx context.Context, k K) (V, error)
		// Set stores a value with optional settings.
		Set(ctx context.Context, k K, v V, opts ...SetOption) error
		// Delete removes a value.
		Delete(ctx context.Context, k K) error
		// Close disconnects from the cache backend.
		Close(ctx context.Context) error
	}
)

// TTL sets the time-to-live for a cache key.
func TTL(ttl time.Duration) SetOption {
	return func(opts *SetOptions) {
		opts.TTL = ttl
	}
}

// Apply applies SetOption functions to SetOptions.
func (opt *SetOptions) Apply(opts ...SetOption) {
	for _, op := range opts {
		op(opt)
	}
}
