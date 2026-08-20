package client

import (
	"context"

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
