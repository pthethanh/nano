package server

import (
	"context"
	"sync"
	"sync/atomic"
)

var (
	def  atomic.Pointer[Server]
	once sync.Once
)

func SetDefault(srv *Server) {
	def.Store(srv)
}

func Default() *Server {
	once.Do(func() {
		if def.Load() != nil {
			return
		}
		def.Store(New())
	})
	return def.Load()
}

func ListenAndServe(ctx context.Context, services ...any) error {
	return Default().ListenAndServe(ctx, services...)
}
