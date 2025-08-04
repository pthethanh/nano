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

// SetDefault sets the default server instance.
func SetDefault(srv *Server) {
	def.Store(srv)
}

// Default returns the default server instance, creating one if needed.
func Default() *Server {
	once.Do(func() {
		if def.Load() != nil {
			return
		}
		def.Store(New())
	})
	return def.Load()
}

// ListenAndServe starts the default server with the provided services.
func ListenAndServe(ctx context.Context, services ...any) error {
	return Default().ListenAndServe(ctx, services...)
}
