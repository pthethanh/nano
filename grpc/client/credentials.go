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
func NewTokenCredentials(token string, secure ...bool) credentials.PerRPCCredentials {
	cred := tokenCredentials{
		token: token,
	}
	if len(secure) > 0 {
		cred.secure = secure[0]
	}
	return cred
}

func (tok tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": tok.token,
	}, nil
}

func (tok tokenCredentials) RequireTransportSecurity() bool {
	return tok.secure
}
