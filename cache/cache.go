package cache

import (
	"context"
	"time"
)

type (
	// SetOptions hold options when setting value for a key.
	SetOptions[K comparable, V any] struct {
		TTL time.Duration
	}

	// SetOption is option when setting value for a key.
	SetOption[K comparable, V any] func(*SetOptions[K, V])

	// Cacher is interface for a cache service.
	Cacher[K comparable, V any] interface {
		// Open establish connection to the target server.
		Open(ctx context.Context) error
		// Get a value, return ErrNotFound if key not found.
		Get(ctx context.Context, k K) (V, error)
		// Set a value
		Set(ctx context.Context, k K, v V, opts ...SetOption[K, V]) error
		// Delete a value
		Delete(ctx context.Context, k K) error
		// Close close the underlying connection.
		Close(ctx context.Context) error
	}
)

// TTL is an option to set Time To Live for a key.
func TTL[K comparable, V any](ttl time.Duration) SetOption[K, V] {
	return func(opts *SetOptions[K, V]) {
		opts.TTL = ttl
	}
}

// Apply apply the options.
func (opt *SetOptions[K, V]) Apply(opts ...SetOption[K, V]) {
	for _, op := range opts {
		op(opt)
	}
}
