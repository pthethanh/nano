// Package memory provides a cache service using in-memory implementation of dgraph-io/ristretto.
package memory

import (
	"context"

	"github.com/dgraph-io/ristretto"
	"github.com/pthethanh/nano/cache"
)

type (
	// Cacher is an in-memory cache implementation using dgraph-io/ristretto cache.
	Cacher[K comparable, V any] struct {
		conf  *ristretto.Config
		cache *ristretto.Cache
	}

	Config = ristretto.Config
)

var (
	// Cacher should implements cache.Cacher
	_ cache.Cacher[string, []byte] = &Cacher[string, []byte]{}
)

func New[K comparable, V any](conf *Config) *Cacher[K, V] {
	return &Cacher[K, V]{
		conf: conf,
	}
}

// Open establish connection to the target server.
func (c *Cacher[K, V]) Open(ctx context.Context) error {
	rc, err := ristretto.NewCache(c.conf)
	if err != nil {
		return err
	}
	c.cache = rc
	return nil
}

// Get a value, return ErrNotFound if key not found.
func (c *Cacher[K, V]) Get(ctx context.Context, k K) (rs V, err error) {
	if err := c.validate(); err != nil {
		return rs, err
	}
	v, ok := c.cache.Get(k)
	if !ok {
		return rs, cache.ErrNotFound
	}
	if vv, ok := v.(V); ok {
		return vv, nil
	}
	return rs, cache.ErrNotFound
}

// Set a value
func (c *Cacher[K, V]) Set(ctx context.Context, k K, v V, opts ...cache.SetOption[K, V]) error {
	if err := c.validate(); err != nil {
		return err
	}
	setOpts := &cache.SetOptions[K, V]{}
	setOpts.Apply(opts...)
	if setOpts.TTL > 0 {
		c.cache.SetWithTTL(k, v, 1, setOpts.TTL)
	} else {
		c.cache.Set(k, v, 1)
	}
	c.cache.Wait()
	return nil
}

// Delete a value
func (c *Cacher[K, V]) Delete(ctx context.Context, k K) error {
	if err := c.validate(); err != nil {
		return err
	}
	c.cache.Del(k)
	return nil
}

// Close close the underlying connection.
func (c *Cacher[K, V]) Close(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return nil
	}
	c.cache.Close()
	return nil
}

func (c *Cacher[K, V]) validate() error {
	if c.cache == nil {
		return cache.ErrInValidConnState
	}
	return nil
}
