package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ContextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context for the server stream.
func (w *ContextServerStream) Context() context.Context {
	return w.ctx
}

// NewContextServerStream returns a ServerStream with the new context.
// If the stream is already a ContextServerStream, its context is updated.
func NewContextServerStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	if existing, ok := stream.(*ContextServerStream); ok {
		existing.ctx = ctx
		return existing
	}
	return &ContextServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}

// IncomingMetadata returns a copy of incoming gRPC metadata from the context.
// It returns nil when the context does not carry incoming metadata.
func IncomingMetadata(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) == 0 {
		return nil
	}
	return md.Copy()
}

// IncomingMetadataValue returns the values for an incoming gRPC metadata key.
func IncomingMetadataValue(ctx context.Context, key string) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) == 0 {
		return nil
	}
	return append([]string(nil), md.Get(key)...)
}
