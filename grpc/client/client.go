package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/health" // enable health check
)

// New creates a client connection to the given target with health check enabled
// and some others default configurations.
func New(_ context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, options...)
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// MustNew same as New but panic if error.
func MustNew(ctx context.Context, address string, options ...grpc.DialOption) *grpc.ClientConn {
	conn, err := New(ctx, address, options...)
	if err != nil {
		panic(err)
	}
	return conn
}
