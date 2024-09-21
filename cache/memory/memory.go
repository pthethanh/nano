// Package memory provides a cache service using in-memory implementation of dgraph-io/ristretto.
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

// Open establish connection to the target server.
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

// Close close the underlying connection.
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
