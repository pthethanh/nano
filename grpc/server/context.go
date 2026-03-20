package server

import (
	"context"
	"fmt"

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

// IncomingMetadataValues returns the values for an incoming gRPC metadata key.
func IncomingMetadataValues(ctx context.Context, key string) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) == 0 {
		return nil
	}
	return append([]string(nil), md.Get(key)...)
}

// IncomingMetadataValue returns the first value for an incoming gRPC metadata key.
// It returns an empty string when the key is not present.
func IncomingMetadataValue(ctx context.Context, key string) string {
	values := IncomingMetadataValues(ctx, key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// RequireIncomingMetadata checks that the provided incoming metadata keys exist.
// It returns an error naming the first missing key.
func RequireIncomingMetadata(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if IncomingMetadataValue(ctx, key) == "" {
			return fmt.Errorf("missing incoming metadata: %s", key)
		}
	}
	return nil
}
