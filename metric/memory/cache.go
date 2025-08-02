package memory

import (
	"strings"
	"sync"
)

type (
	cache[T any] struct {
		cache *sync.Map
	}
	key struct {
		k  string
		vs string
	}
)

func newCache[T any]() *cache[T] {
	return &cache[T]{
		cache: new(sync.Map),
	}
}

func (c *cache[T]) loadOrCreate(k string, vs []string, create func() T) T {
	mk := key{k: k, vs: strings.Join(vs, "|")}
	if v, ok := c.cache.Load(mk); ok {
		return v.(T)
	}
	v, _ := c.cache.LoadOrStore(mk, create())
	return v.(T)
}
