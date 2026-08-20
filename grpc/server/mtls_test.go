package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/client"
	"github.com/pthethanh/nano/grpc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func credentialsFromPool(pool *x509.CertPool) credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{RootCAs: pool})
}

// testCA is a minimal self-signed certificate authority for mTLS tests.
type testCA struct {
	certPEM []byte
	certFile,
	keyFile string
	cert *x509.Certificate
	key  *rsa.PrivateKey
}

func newTestCA(t *testing.T) *testCA {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	certFile := writePEM(t, "ca-cert.pem", certPEM)
	return &testCA{certPEM: certPEM, certFile: certFile, cert: cert, key: key}
}

// issue signs a new leaf certificate for commonName using this CA.
func (ca *testCA) issue(t *testing.T, commonName string, extKeyUsage x509.ExtKeyUsage) (certFile, keyFile string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{extKeyUsage},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return writePEM(t, commonName+"-cert.pem", certPEM), writePEM(t, commonName+"-key.pem", keyPEM)
}

func writePEM(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := t.TempDir() + "/" + name
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// testEchoServiceDesc is a minimal hand-rolled unary service used to prove
// an mTLS connection can carry a real RPC end-to-end, and to observe the
// peer certificate seen server-side.
var testEchoServiceDesc = &grpc.ServiceDesc{
	ServiceName: "test.Echo",
	HandlerType: (*any)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(wrapperspb.StringValue)
				if err := dec(in); err != nil {
					return nil, err
				}
				handler := func(ctx context.Context, req any) (any, error) {
					cn := ""
					if cert, ok := server.PeerCertificate(ctx); ok {
						cn = cert.Subject.CommonName
					}
					return wrapperspb.String(req.(*wrapperspb.StringValue).Value + "|" + cn), nil
				}
				if interceptor == nil {
					return handler(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{FullMethod: "/test.Echo/Echo"}, handler)
			},
		},
	},
}

func TestMutualTLS_ClientWithValidCertCanCallSeparateGRPCListener(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey := ca.issue(t, "server", x509.ExtKeyUsageServerAuth)
	clientCert, clientKey := ca.issue(t, "test-client", x509.ExtKeyUsageClientAuth)

	grpcLis := mustListen(t)
	httpLis := mustListen(t)
	t.Cleanup(func() { _ = grpcLis.Close(); _ = httpLis.Close() })

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	srv := server.New(
		server.SeparateListeners(grpcLis, httpLis),
		server.MutualTLS(serverCert, serverKey, ca.certFile),
	)
	srv.RegisterService(testEchoServiceDesc, nil)

	go func() { _ = srv.ListenAndServe(ctx) }()
	t.Cleanup(func() { cancel(nil) })

	dialCreds, err := client.WithMutualTLS(clientCert, clientKey, ca.certFile)
	if err != nil {
		t.Fatalf("WithMutualTLS() error = %v", err)
	}
	conn, err := grpc.NewClient(grpcLis.Addr().String(), dialCreds)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer conn.Close()

	callCtx, callCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer callCancel()
	out := new(wrapperspb.StringValue)
	if err := conn.Invoke(callCtx, "/test.Echo/Echo", wrapperspb.String("hi"), out); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if want := "hi|test-client"; out.Value != want {
		t.Errorf("got response=%q, want %q (peer cert CommonName must reach the handler via server.PeerCertificate)", out.Value, want)
	}
}

func TestMutualTLS_ClientWithoutCertIsRejected(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey := ca.issue(t, "server", x509.ExtKeyUsageServerAuth)

	grpcLis := mustListen(t)
	httpLis := mustListen(t)
	t.Cleanup(func() { _ = grpcLis.Close(); _ = httpLis.Close() })

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	srv := server.New(
		server.SeparateListeners(grpcLis, httpLis),
		server.MutualTLS(serverCert, serverKey, ca.certFile),
	)
	srv.RegisterService(testEchoServiceDesc, nil)

	go func() { _ = srv.ListenAndServe(ctx) }()
	t.Cleanup(func() { cancel(nil) })

	// Plain server-only TLS trust, no client certificate presented.
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca.certPEM)
	conn, err := grpc.NewClient(grpcLis.Addr().String(), grpc.WithTransportCredentials(credentialsFromPool(pool)))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer conn.Close()

	assertHandshakeFails(t, conn)
}

func TestMutualTLS_ClientWithWrongCACertIsRejected(t *testing.T) {
	ca := newTestCA(t)
	otherCA := newTestCA(t)
	serverCert, serverKey := ca.issue(t, "server", x509.ExtKeyUsageServerAuth)
	wrongClientCert, wrongClientKey := otherCA.issue(t, "impostor", x509.ExtKeyUsageClientAuth)

	grpcLis := mustListen(t)
	httpLis := mustListen(t)
	t.Cleanup(func() { _ = grpcLis.Close(); _ = httpLis.Close() })

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	srv := server.New(
		server.SeparateListeners(grpcLis, httpLis),
		server.MutualTLS(serverCert, serverKey, ca.certFile),
	)
	srv.RegisterService(testEchoServiceDesc, nil)

	go func() { _ = srv.ListenAndServe(ctx) }()
	t.Cleanup(func() { cancel(nil) })

	dialCreds, err := client.WithMutualTLS(wrongClientCert, wrongClientKey, ca.certFile)
	if err != nil {
		t.Fatalf("WithMutualTLS() error = %v", err)
	}
	conn, err := grpc.NewClient(grpcLis.Addr().String(), dialCreds)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer conn.Close()

	assertHandshakeFails(t, conn)
}

func assertHandshakeFails(t *testing.T, conn *grpc.ClientConn) {
	t.Helper()
	callCtx, callCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer callCancel()
	out := new(wrapperspb.StringValue)
	err := conn.Invoke(callCtx, "/test.Echo/Echo", wrapperspb.String("hi"), out)
	if err == nil {
		t.Fatal("expected the call to fail (TLS handshake should have been rejected)")
	}
}
