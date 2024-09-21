package cache

import (
	"context"
	"time"
)

type (
	// SetOptions hold options when setting value for a key.
	SetOptions struct {
		TTL time.Duration
	}

	// SetOption is option when setting value for a key.
	SetOption func(*SetOptions)

	// Cacher is interface for a cache service.
	Cacher[K comparable, V any] interface {
		// Open establish connection to the target server.
		Open(ctx context.Context) error
		// Get a value, return ErrNotFound if key not found.
		Get(ctx context.Context, k K) (V, error)
		// Set a value
		Set(ctx context.Context, k K, v V, opts ...SetOption) error
		// Delete a value
		Delete(ctx context.Context, k K) error
		// Close close the underlying connection.
		Close(ctx context.Context) error
	}
)

// TTL is an option to set Time To Live for a key.
func TTL(ttl time.Duration) SetOption {
	return func(opts *SetOptions) {
		opts.TTL = ttl
	}
}

// Apply apply the options.
func (opt *SetOptions) Apply(opts ...SetOption) {
	for _, op := range opts {
		op(opt)
	}
}
