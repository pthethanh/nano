package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	server "github.com/pthethanh/nano/grpc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testHTTPService struct {
	prefix string
	body   string
}

func (h testHTTPService) HTTPHandler() (string, http.Handler) {
	return h.prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, h.body)
	})
}

type testDescService interface {
	testDescService()
}

type testDescServiceImpl struct{}

func (*testDescServiceImpl) testDescService() {}

func (*testDescServiceImpl) ServiceDesc() *grpc.ServiceDesc {
	return &grpc.ServiceDesc{
		ServiceName: "test.DescService",
		HandlerType: (*testDescService)(nil),
	}
}

func TestIncomingHeaderMatchers(t *testing.T) {
	mux := runtime.NewServeMux(server.WithIncomingHeaderMatcher([]string{"X-Test", "X-Test"}))
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("x-test", "value")
	req.Header.Set("X-Request-Id", "req-1")

	ctx, err := runtime.AnnotateIncomingContext(context.Background(), mux, req, "/svc/method")
	if err != nil {
		t.Fatal(err)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		t.Fatal("expected incoming metadata")
	}
	if got, want := md.Get("X-Test"), []string{"value"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("got X-Test=%v, want %v", got, want)
	}
	if got, want := md.Get("X-Request-Id"), []string{"req-1"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("got X-Request-Id=%v, want %v", got, want)
	}

	prefixMux := runtime.NewServeMux(server.WithIncomingHeaderPrefixMatcher([]string{"x-app", "x-app"}))
	prefixReq := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	prefixReq.Header.Set("x-app-tenant", "tenant-1")
	prefixReq.Header.Set("Api-Key", "secret")

	ctx, err = runtime.AnnotateIncomingContext(context.Background(), prefixMux, prefixReq, "/svc/method")
	if err != nil {
		t.Fatal(err)
	}
	md, ok = metadata.FromIncomingContext(ctx)
	if !ok {
		t.Fatal("expected incoming metadata")
	}
	if got, want := md.Get("X-App-Tenant"), []string{"tenant-1"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("got X-App-Tenant=%v, want %v", got, want)
	}
	if got, want := md.Get("Api-Key"), []string{"secret"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("got Api-Key=%v, want %v", got, want)
	}
}

func TestIncomingHeaderPrefixMatcherFallsBackToDefault(t *testing.T) {
	mux := runtime.NewServeMux(server.WithIncomingHeaderPrefixMatcher([]string{"x-app"}))
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Unmatched", "value")

	ctx, err := runtime.AnnotateIncomingContext(context.Background(), mux, req, "/svc/method")
	if err != nil {
		t.Fatal(err)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		t.Fatal("expected incoming metadata")
	}
	if got := md.Get("X-Unmatched"); len(got) != 0 {
		t.Fatalf("got X-Unmatched=%v, want none", got)
	}
}

func TestNewWithTLSExposesHTTPSAddress(t *testing.T) {
	certFile, keyFile := writeTestCertPair(t)
	srv := server.New(
		server.Address("127.0.0.1:8080"),
		server.Timeout(time.Second, 2*time.Second),
		server.TLS(certFile, keyFile),
	)

	if got, want := srv.Address(), "127.0.0.1:8080"; got != want {
		t.Fatalf("got address=%q, want %q", got, want)
	}
	if got, want := srv.HTTPAddress(), "https://127.0.0.1:8080"; got != want {
		t.Fatalf("got http address=%q, want %q", got, want)
	}
	if len(srv.DialOpts()) == 0 {
		t.Fatal("expected dial options")
	}
}

func TestListenAndServeServesHTTPHandlerWithMiddleware(t *testing.T) {
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
		server.Middlewares(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Middleware", "applied")
				next.ServeHTTP(w, r)
			})
		}),
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
	if got, want := resp.Header.Get("X-Middleware"), "applied"; got != want {
		t.Fatalf("got header=%q, want %q", got, want)
	}

	stopErr := errors.New("stop server")
	cancel(stopErr)
	if err := <-errCh; !errors.Is(err, stopErr) && !errors.Is(err, context.Canceled) {
		t.Fatalf("got err=%v, want shutdown cancellation", err)
	}
}

func TestListenAndServeUsesNotFoundHandler(t *testing.T) {
	grpcLis := mustListen(t)
	httpLis := mustListen(t)
	t.Cleanup(func() {
		_ = grpcLis.Close()
		_ = httpLis.Close()
	})

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusTeapot)
	})
	srv := server.New(
		server.SeparateListeners(grpcLis, httpLis),
		server.NotFoundHandler(notFound),
	)
	srv.RegisterService((&testDescServiceImpl{}).ServiceDesc(), &testDescServiceImpl{})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, testHTTPService{prefix: "/hello", body: "world"})
	}()

	resp := waitForHTTP(t, "http://"+httpLis.Addr().String()+"/missing")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := resp.StatusCode, http.StatusTeapot; got != want {
		t.Fatalf("got status=%d, want %d", got, want)
	}
	if got, want := string(body), "missing\n"; got != want {
		t.Fatalf("got body=%q, want %q", got, want)
	}

	stopErr := errors.New("stop server")
	cancel(stopErr)
	if err := <-errCh; !errors.Is(err, stopErr) && !errors.Is(err, context.Canceled) {
		t.Fatalf("got err=%v, want shutdown cancellation", err)
	}
}

func TestListenAndServeRejectsUnknownServiceType(t *testing.T) {
	srv := server.New(server.Address("127.0.0.1:8080"))
	if err := srv.ListenAndServe(context.Background(), struct{}{}); !errors.Is(err, server.ErrUnknownServiceType) {
		t.Fatalf("got err=%v, want %v", err, server.ErrUnknownServiceType)
	}
}

func TestTLSPanicsOnInvalidFiles(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = server.TLS("missing-cert.pem", "missing-key.pem")
}

func TestDefaultFunctions(t *testing.T) {
	old := server.Default()
	t.Cleanup(func() {
		server.SetDefault(old)
	})

	custom := server.New(server.Address("127.0.0.1:8080"))
	server.SetDefault(custom)
	if got := server.Default(); got != custom {
		t.Fatal("expected stored default server")
	}

	server.SetDefault(server.New(server.Address("bad-address")))
	if err := server.ListenAndServe(context.Background()); err == nil {
		t.Fatal("expected top-level listen error")
	}
}

func mustListen(t *testing.T) net.Listener {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return lis
}

func waitForHTTP(t *testing.T, url string) *http.Response {
	t.Helper()
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			return resp
		}
		if time.Now().After(deadline) {
			t.Fatalf("request to %s did not succeed: %v", url, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func writeTestCertPair(t *testing.T) (string, string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "server.test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"server.test"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	certFile := t.TempDir() + "/cert.pem"
	keyFile := t.TempDir() + "/key.pem"
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}
