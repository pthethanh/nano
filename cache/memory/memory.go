package memory

import (
	"context"
	"errors"

	"github.com/coocood/freecache"
	"github.com/pthethanh/nano/cache"
)

type (
	// Memory is an implementation of cache.Cacher
	Memory struct {
		cache *freecache.Cache
	}

	Option func(*Memory)
)

var (
	_ cache.Cacher = (*Memory)(nil)

	// ErrInvalidConnectionState indicate that the connection has not been opened properly.
	ErrInvalidConnectionState = errors.New("invalid connection state")
)

// New return new memory cache.
func New(size int) *Memory {
	m := &Memory{
		cache: freecache.NewCache(size),
	}
	return m
}

// Get a value.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := m.cache.Get([]byte(key))
	if errors.Is(err, freecache.ErrNotFound) {
		return nil, cache.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Set a value.
func (m *Memory) Set(ctx context.Context, key string, val []byte, opts ...cache.SetOption) error {
	opt := &cache.SetOptions{}
	opt.Apply(opts...)
	ttl := opt.TTL.Seconds()
	if ttl < 1 && ttl > 0 {
		ttl = 1 // at least 1 sec if defined.
	}
	if err := m.cache.Set([]byte(key), val, int(ttl)); err != nil {
		return err
	}
	return nil
}

// Delete a value.
func (m *Memory) Delete(ctx context.Context, key string) error {
	m.cache.Del([]byte(key))
	return nil
}

// Open make the cacher ready for using.
func (m *Memory) Open(ctx context.Context) error {
	// nothing to do
	return nil
}

// Close close underlying resources.
func (m *Memory) Close(ctx context.Context) error {
	m.cache.Clear()
	return nil
}

// CheckHealth return health check func.
func (m *Memory) CheckHealth(ctx context.Context) error {
	return nil
}
