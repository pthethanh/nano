// Package memory provides a cache service using an in-memory implementation
// backed by jellydator/ttlcache.
package memory

import (
	"context"

	"github.com/jellydator/ttlcache/v3"
	"github.com/pthethanh/nano/cache"
)

type (
	// Cacher is an in-memory cache implementation.
	Cacher[K comparable, V any] struct {
		cache *ttlcache.Cache[K, V]
	}
)

var (
	// Cacher should implements cache.Cacher
	_ cache.Cacher[string, []byte] = &Cacher[string, []byte]{}
)

func New[K comparable, V any](opts ...ttlcache.Option[K, V]) *Cacher[K, V] {
	return &Cacher[K, V]{
		cache: ttlcache.New[K, V](opts...),
	}
}

// Open starts the background sweep goroutine that proactively evicts
// expired entries. It is not required before calling Get/Set/Delete: the
// cache is fully usable immediately after New, and Get already treats an
// expired entry as not found even if it hasn't been swept yet. Calling Open
// more than once is safe (ttlcache.Start is idempotent while already
// running).
func (c *Cacher[K, V]) Open(ctx context.Context) error {
	go c.cache.Start()
	return nil
}

// Get a value, return ErrNotFound if key not found.
func (c *Cacher[K, V]) Get(ctx context.Context, k K) (rs V, err error) {
	if err := c.validate(); err != nil {
		return rs, err
	}
	if item := c.cache.Get(k); item != nil {
		return item.Value(), nil
	}
	return rs, cache.ErrNotFound
}

// Set a value
func (c *Cacher[K, V]) Set(ctx context.Context, k K, v V, opts ...cache.SetOption) error {
	if err := c.validate(); err != nil {
		return err
	}
	setOpts := &cache.SetOptions{}
	setOpts.Apply(opts...)
	if setOpts.TTL > 0 {
		c.cache.Set(k, v, setOpts.TTL)
	} else {
		c.cache.Set(k, v, ttlcache.NoTTL)
	}
	return nil
}

// Delete a value
func (c *Cacher[K, V]) Delete(ctx context.Context, k K) error {
	if err := c.validate(); err != nil {
		return err
	}
	c.cache.Delete(k)
	return nil
}

// Close stops the background sweep goroutine and clears all entries.
// Unlike Get/Set/Delete, Close deliberately ignores validate()'s error and
// always returns nil: closing an already-invalid or never-opened Cacher is
// treated as already-closed rather than an error, so callers can safely
// defer Close() without checking whether Open() ever ran or succeeded.
func (c *Cacher[K, V]) Close(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return nil
	}
	c.cache.Stop()
	c.cache.DeleteAll()
	return nil
}

func (c *Cacher[K, V]) validate() error {
	if c.cache == nil {
		return cache.ErrInValidConnState
	}
	return nil
}
