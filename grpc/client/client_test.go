package client_test

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
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestMustNewReturnsConnection(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	healthpb.RegisterHealthServer(srv, health.NewServer())
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	conn := client.MustNew(t.Context(), "passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	t.Cleanup(func() { _ = conn.Close() })

	client := healthpb.NewHealthClient(conn)
	if _, err := client.Check(t.Context(), &healthpb.HealthCheckRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestNewWithExplicitInsecureTransportCredentials(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	healthpb.RegisterHealthServer(srv, health.NewServer())
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	conn, err := client.New(t.Context(), "passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := healthpb.NewHealthClient(conn)
	if _, err := client.Check(t.Context(), &healthpb.HealthCheckRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestNewWithExplicitTLSCredentials(t *testing.T) {
	serverTLS, clientTLS := newTLSConfigPair(t)
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(serverTLS)))
	healthpb.RegisterHealthServer(srv, health.NewServer())
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	conn, err := client.New(t.Context(), "passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(credentials.NewTLS(clientTLS)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := healthpb.NewHealthClient(conn)
	if _, err := client.Check(t.Context(), &healthpb.HealthCheckRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestNewTokenCredentials(t *testing.T) {
	cred := client.NewTokenCredentials("Bearer abc")
	md, err := cred.GetRequestMetadata(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := md["authorization"]; got != "Bearer abc" {
		t.Fatalf("got authorization=%q, want %q", got, "Bearer abc")
	}
	if cred.RequireTransportSecurity() {
		t.Fatal("got RequireTransportSecurity=true, want false")
	}

	secureCred := client.NewTokenCredentials("Bearer abc", true)
	if !secureCred.RequireTransportSecurity() {
		t.Fatal("got RequireTransportSecurity=false, want true")
	}
}

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}
}

func newTLSConfigPair(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "bufnet.test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"bufnet.test"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatal("failed to append root cert")
	}

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	clientTLS := &tls.Config{
		RootCAs:    pool,
		ServerName: "bufnet.test",
	}
	return serverTLS, clientTLS
}
