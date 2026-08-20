package server

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const defaultTestBufSize = 1024 * 1024

// NewTest starts a Server on an in-memory listener, registers services on it,
// and returns a ready-to-use client connection. opts are passed through to New;
// Listener is configured automatically and must not be supplied by the caller.
// The server and connection are stopped via t.Cleanup.
//
//	conn := server.NewTest(t, []any{&myServiceImpl{}})
//	client := mypb.NewMyServiceClient(conn)
func NewTest(t testing.TB, services []any, opts ...grpc.ServerOption) *grpc.ClientConn {
	t.Helper()

	lis := bufconn.Listen(defaultTestBufSize)
	srv := New(append([]grpc.ServerOption{Listener(lis)}, opts...)...)

	ctx, cancel := context.WithCancelCause(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, services...)
	}()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		cancel(nil)
		t.Fatalf("server.NewTest: dial failed: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		cancel(nil)
		<-errCh
	})

	return conn
}
