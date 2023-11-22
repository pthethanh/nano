package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/health" // enable health check
)

// Dial creates a client connection to the given target with health check enabled
// and some others default configurations.
func Dial(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, options...)
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// MustDial same as Dial but panic if error.
func MustDial(ctx context.Context, address string, options ...grpc.DialOption) *grpc.ClientConn {
	conn, err := Dial(ctx, address, options...)
	if err != nil {
		panic(err)
	}
	return conn
}
