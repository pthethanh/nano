package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/health" // enable health check
)

// New creates a gRPC client connection to address.
//
// By default it uses insecure transport credentials unless explicit dial options
// override them. Import health checking is enabled so generated clients can use
// gRPC health checks without extra setup.
func New(_ context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, options...)
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// MustNew is like New but panics if the connection cannot be created.
func MustNew(ctx context.Context, address string, options ...grpc.DialOption) *grpc.ClientConn {
	conn, err := New(ctx, address, options...)
	if err != nil {
		panic(err)
	}
	return conn
}
