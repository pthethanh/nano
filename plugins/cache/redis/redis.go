// Package redis provides a Redis-backed cache implementation.
package redis

import (
	"context"
	"time"

	"github.com/pthethanh/nano/cache"
	goredis "github.com/redis/go-redis/v9"
)

type Cacher[K comparable, V any] struct {
	client  goredis.UniversalClient
	opts    *goredis.UniversalOptions
	managed bool
	keyFunc KeyFunc[K]
	codec   Codec[V]
}

var _ cache.Cacher[string, []byte] = (*Cacher[string, []byte])(nil)

func New[K comparable, V any](opts ...Option[K, V]) *Cacher[K, V] {
	c := &Cacher[K, V]{
		opts: &goredis.UniversalOptions{
			Addrs: []string{"127.0.0.1:6379"},
		},
		managed: true,
		keyFunc: DefaultKeyFunc[K],
		codec:   JSONCodec[V]{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Open establishes the connection to Redis.
func (c *Cacher[K, V]) Open(ctx context.Context) error {
	if c.client != nil {
		return c.client.Ping(ctx).Err()
	}
	c.client = goredis.NewUniversalClient(c.opts)
	if err := c.client.Ping(ctx).Err(); err != nil {
		_ = c.client.Close()
		c.client = nil
		return err
	}
	c.managed = true
	return nil
}

// Get retrieves a value and returns cache.ErrNotFound when the key does not exist.
func (c *Cacher[K, V]) Get(ctx context.Context, k K) (rs V, err error) {
	if err := c.validate(); err != nil {
		return rs, err
	}
	val, err := c.client.Get(ctx, c.keyFunc(k)).Bytes()
	if err == goredis.Nil {
		return rs, cache.ErrNotFound
	}
	if err != nil {
		return rs, err
	}
	if err := c.codec.Unmarshal(val, &rs); err != nil {
		return rs, err
	}
	return rs, nil
}

// Set stores a value with optional TTL.
func (c *Cacher[K, V]) Set(ctx context.Context, k K, v V, opts ...cache.SetOption) error {
	if err := c.validate(); err != nil {
		return err
	}
	setOpts := &cache.SetOptions{}
	setOpts.Apply(opts...)
	b, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.keyFunc(k), b, ttl(setOpts.TTL)).Err()
}

// Delete removes a value.
func (c *Cacher[K, V]) Delete(ctx context.Context, k K) error {
	if err := c.validate(); err != nil {
		return err
	}
	return c.client.Del(ctx, c.keyFunc(k)).Err()
}

// Close closes the managed Redis client.
func (c *Cacher[K, V]) Close(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return nil
	}
	if !c.managed {
		c.client = nil
		return nil
	}
	err := c.client.Close()
	c.client = nil
	return err
}

func (c *Cacher[K, V]) validate() error {
	if c.client == nil {
		return cache.ErrInValidConnState
	}
	return nil
}

func ttl(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return d
}
