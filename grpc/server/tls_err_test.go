package server_test

import (
	"testing"

	"github.com/pthethanh/nano/grpc/server"
)

func TestTLSErr_ReturnsErrorInsteadOfPanickingOnBadFiles(t *testing.T) {
	_, err := server.TLSErr("missing-cert.pem", "missing-key.pem")
	if err == nil {
		t.Fatal("expected an error for a missing cert file, got nil")
	}
}

func TestTLSErr_ReturnsUsableOptionForValidFiles(t *testing.T) {
	certFile, keyFile := writeTestCertPair(t)

	opt, err := server.TLSErr(certFile, keyFile)
	if err != nil {
		t.Fatalf("TLSErr() error = %v", err)
	}
	if opt == nil {
		t.Fatal("expected a non-nil grpc.ServerOption")
	}
	// Must not panic when applied.
	_ = server.New(opt)
}
