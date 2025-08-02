package memory

import (
	"strings"
	"sync"
)

type (
	cache[T any] struct {
		cache map[key]T
		mux   *sync.Mutex
	}
	key struct {
		k  string
		vs string
	}
)

func newCache[T any]() *cache[T] {
	return &cache[T]{
		cache: make(map[key]T),
		mux:   new(sync.Mutex),
	}
}

func (c *cache[T]) loadOrCreate(k string, vs []string, create func() T) T {
	c.mux.Lock()
	defer c.mux.Unlock()
	mk := key{k: k, vs: strings.Join(vs, "|")}
	if v, ok := c.cache[mk]; ok {
		return v
	}
	nc := create()
	c.cache[mk] = nc
	return nc
}
