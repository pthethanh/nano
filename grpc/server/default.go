package server

import (
	"context"
	"sync/atomic"
)

var def atomic.Pointer[Server]

// SetDefault sets the default server instance.
func SetDefault(srv *Server) {
	def.Store(srv)
}

// Default returns the default server instance, creating one if needed.
//
// Safe for concurrent use, including concurrently with SetDefault: at most
// one lazily-constructed Server is ever installed, and it never overwrites
// a Server set by a concurrent SetDefault call.
func Default() *Server {
	if srv := def.Load(); srv != nil {
		return srv
	}
	srv := New()
	if !def.CompareAndSwap(nil, srv) {
		// Someone else (another concurrent Default() or a concurrent
		// SetDefault) already installed one; use that instead of the one
		// we just built.
		return def.Load()
	}
	return srv
}

// ListenAndServe starts the default server with the provided services.
func ListenAndServe(ctx context.Context, services ...any) error {
	return Default().ListenAndServe(ctx, services...)
}
