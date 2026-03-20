package client

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// OutgoingMetadata sets outgoing gRPC metadata on the context.
// Existing outgoing metadata is preserved, except matching keys are replaced.
func OutgoingMetadata(ctx context.Context, pairs ...string) context.Context {
	if len(pairs) == 0 {
		return ctx
	}

	md := metadata.Pairs(pairs...)
	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(outgoing) == 0 {
		return metadata.NewOutgoingContext(ctx, md)
	}

	merged := outgoing.Copy()
	for key, values := range md {
		merged[key] = append([]string(nil), values...)
	}
	return metadata.NewOutgoingContext(ctx, merged)
}

// AppendOutgoingMetadata appends outgoing gRPC metadata to the context.
func AppendOutgoingMetadata(ctx context.Context, pairs ...string) context.Context {
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

// ForwardIncomingMetadata forwards incoming gRPC metadata onto the outgoing context.
// When keys are provided, only those incoming metadata keys are forwarded.
// Existing outgoing metadata is preserved, except matching keys are replaced
// with the forwarded incoming values.
func ForwardIncomingMetadata(ctx context.Context, keys ...string) context.Context {
	incoming, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(incoming) == 0 {
		return ctx
	}

	forwarded := incoming.Copy()
	if len(keys) > 0 {
		forwarded = metadata.MD{}
		for _, key := range keys {
			key = strings.ToLower(key)
			values, ok := incoming[key]
			if !ok {
				continue
			}
			forwarded[key] = append([]string(nil), values...)
		}
		if len(forwarded) == 0 {
			return ctx
		}
	}

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(outgoing) == 0 {
		return metadata.NewOutgoingContext(ctx, forwarded)
	}

	merged := outgoing.Copy()
	for key, values := range forwarded {
		merged[key] = append([]string(nil), values...)
	}
	return metadata.NewOutgoingContext(ctx, merged)
}
