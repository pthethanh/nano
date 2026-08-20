package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type tokenCredentials struct {
	token  string
	secure bool
}

// NewTokenCredentials returns a PerRPCCredentials using the provided token.
// secure controls RequireTransportSecurity: pass true unless the connection
// is already known to be secured some other way (e.g. mTLS at a lower
// layer, or a trusted local/internal network).
func NewTokenCredentials(token string, secure bool) credentials.PerRPCCredentials {
	return tokenCredentials{
		token:  token,
		secure: secure,
	}
}

func (tok tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": tok.token,
	}, nil
}

func (tok tokenCredentials) RequireTransportSecurity() bool {
	return tok.secure
}

// MutualTLSCredentials loads client mTLS transport credentials: certFile
// and keyFile identify this client to the server, and caFile is the CA
// (PEM-encoded) used to verify the server's certificate.
func MutualTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("client: no certificates found in CA file %s", caFile)
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}), nil
}

// WithMutualTLS returns a DialOption using mutual TLS: certFile and keyFile
// identify this client to the server, and caFile is the CA used to verify
// the server's certificate.
func WithMutualTLS(certFile, keyFile, caFile string) (grpc.DialOption, error) {
	creds, err := MutualTLSCredentials(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	return grpc.WithTransportCredentials(creds), nil
}
