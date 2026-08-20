package server_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/pthethanh/nano/grpc/server"
)

// TestKeepaliveDefaults_ServerStartsAndServesNormally proves that adding
// KeepaliveDefaults() to a server's options doesn't break normal startup or
// request handling (the underlying grpc.KeepaliveParams/EnforcementPolicy
// ServerOptions are wired correctly into grpc.NewServer).
func TestKeepaliveDefaults_ServerStartsAndServesNormally(t *testing.T) {
	grpcLis := mustListen(t)
	httpLis := mustListen(t)
	t.Cleanup(func() {
		_ = grpcLis.Close()
		_ = httpLis.Close()
	})

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	srv := server.New(
		server.SeparateListeners(grpcLis, httpLis),
		server.KeepaliveDefaults(),
	)
	srv.RegisterService((&testDescServiceImpl{}).ServiceDesc(), &testDescServiceImpl{})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, testHTTPService{prefix: "/hello", body: "world"})
	}()

	resp := waitForHTTP(t, "http://"+httpLis.Addr().String()+"/hello")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), "world"; got != want {
		t.Fatalf("got body=%q, want %q", got, want)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status=%d, want 200", resp.StatusCode)
	}

	stopErr := errors.New("stop server")
	cancel(stopErr)
	if err := <-errCh; !errors.Is(err, stopErr) && !errors.Is(err, context.Canceled) {
		t.Fatalf("got err=%v, want shutdown cancellation", err)
	}
}
