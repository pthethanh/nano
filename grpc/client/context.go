package client

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// NewContext return new out going context with metadata copied
// from incoming or outgoing context all together.
// It also attach a X-Request-Id for tracing into the context if not present yet.
// Outgoing context takes higher priority than incoming context.
func NewContext(ctx context.Context) context.Context {
	md1, _ := metadata.FromOutgoingContext(ctx)
	md2, _ := metadata.FromIncomingContext(ctx)
	md := metadata.Join(md1, md2)
	if len(md["X-Request-Id"]) == 0 {
		metadata.Pairs("X-Request-Id", uuid.NewString())
	}
	return metadata.NewOutgoingContext(ctx, md)
}
