package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type (
	middleware = func(http.Handler) http.Handler

	customServerOption interface {
		nanoCustomOpt()
	}

	emptyOpt struct {
		grpc.ServerOption
		customServerOption
	}

	loggerOpt struct {
		emptyOpt
		logger logger
	}

	onShutdownOpt struct {
		emptyOpt
		f func()
	}
	gwOpt struct {
		emptyOpt
		opts []runtime.ServeMuxOption
	}

	timeoutOpt struct {
		emptyOpt
		read  time.Duration
		write time.Duration
	}

	tlsOpt struct {
		emptyOpt
		certFile   string
		keyFile    string
		dialOpt    []grpc.DialOption
		grpcCreds  credentials.TransportCredentials
		clientAuth tls.ClientAuthType
		clientCAs  *x509.CertPool
	}

	addrOpt struct {
		emptyOpt
		grpcAddr string
		httpAddr string
	}

	apiPrefixOpt struct {
		emptyOpt
		prefix string
	}

	handlerOpt struct {
		emptyOpt
		prefix string
		h      http.Handler
	}

	notFoundHandlerOpt struct {
		emptyOpt
		h http.Handler
	}
	mdwOpt struct {
		emptyOpt
		mdws []middleware
	}

	lisOpt struct {
		emptyOpt
		lis     net.Listener
		httpLis net.Listener
	}

	shutdownTimeout struct {
		emptyOpt
		timeout time.Duration
	}

	autoMaxProcsOpt struct {
		emptyOpt
	}

	keepaliveOpt struct {
		emptyOpt
		params keepalive.ServerParameters
		policy keepalive.EnforcementPolicy
	}
)

var (
	defaultGWPassthroughHeaders = []string{"X-Request-Id", "X-Correlation-ID", "Api-Key"}
)

// Logger replaces the server logger used for lifecycle and registration logs.
func Logger(logger logger) grpc.ServerOption {
	return loggerOpt{
		logger: logger,
	}
}

// OnShutdown registers a callback that runs before the server shuts down.
func OnShutdown(f func()) grpc.ServerOption {
	return onShutdownOpt{
		f: f,
	}
}

// GateWayOpts appends raw grpc-gateway ServeMux options to the internal gateway mux.
//
// Use this when the built-in helpers in this package are not enough and you need
// direct control over grpc-gateway behavior.
func GateWayOpts(opts ...runtime.ServeMuxOption) grpc.ServerOption {
	return gwOpt{
		opts: opts,
	}
}

// GatewayForwardHeaders forwards the provided HTTP header names to gRPC metadata.
//
// The configured headers are forwarded in addition to the default passthrough
// headers used by this package.
func GatewayForwardHeaders(keys ...string) grpc.ServerOption {
	return GateWayOpts(WithIncomingHeaderMatcher(keys))
}

// GatewayForwardHeadersByPrefix forwards HTTP headers whose canonicalized names
// start with one of the provided prefixes.
//
// This is useful for families of headers such as `X-Forwarded-` or custom
// tracing and tenant headers.
func GatewayForwardHeadersByPrefix(prefixes ...string) grpc.ServerOption {
	return GateWayOpts(WithIncomingHeaderPrefixMatcher(prefixes))
}

// Timeout sets read and write timeouts for the internal HTTP server.
//
// These timeouts apply to HTTP and grpc-gateway traffic handled by the embedded
// HTTP server. They do not change gRPC per-request deadlines.
func Timeout(read, write time.Duration) grpc.ServerOption {
	return timeoutOpt{
		read:  read,
		write: write,
	}
}

// TLS enables TLS using the provided certificate and key files.
//
// It also updates the server's self-dial options so the internal gateway dials
// the gRPC server using TLS instead of insecure credentials.
//
// Panics if the certificate file can't be loaded. grpc.ServerOption values
// are typically constructed inline (e.g. server.New(server.TLS(...))),
// which offers no path to return an error, hence the panic; use TLSErr if
// you need to validate the cert path without crashing the process.
func TLS(certFile, keyFile string) grpc.ServerOption {
	opt, err := TLSErr(certFile, keyFile)
	if err != nil {
		panic(err)
	}
	return opt
}

// TLSErr is the error-returning equivalent of TLS, for callers that load
// certificate paths from configuration and want to handle a bad path as a
// normal error instead of a panic.
func TLSErr(certFile, keyFile string) (grpc.ServerOption, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	dialCreds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		return nil, err
	}
	return tlsOpt{
		keyFile:  keyFile,
		certFile: certFile,
		dialOpt:  []grpc.DialOption{grpc.WithTransportCredentials(dialCreds)},
		// Also used for a gRPC listener started via SeparateAddresses /
		// SeparateListeners, which bypasses the HTTP server's ServeTLS
		// entirely and needs its own transport credentials.
		grpcCreds: credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}}),
	}, nil
}

// MutualTLS enables mutual TLS: certFile/keyFile identify this server, and
// clientCAFile is the CA (PEM-encoded) used to verify client certificates.
// Only clients presenting a certificate signed by that CA are accepted.
//
// It also updates the server's self-dial options so the internal gateway
// presents certFile/keyFile as its own client identity and trusts
// clientCAFile to verify the gRPC server's certificate — this assumes the
// same CA signs both the server certificate and client certificates, which
// is the common case for a private internal mTLS setup.
//
// Panics if the certificate or CA file can't be loaded; use MutualTLSErr to
// handle that as a normal error instead.
func MutualTLS(certFile, keyFile, clientCAFile string) grpc.ServerOption {
	opt, err := MutualTLSErr(certFile, keyFile, clientCAFile)
	if err != nil {
		panic(err)
	}
	return opt
}

// MutualTLSErr is the error-returning equivalent of MutualTLS.
func MutualTLSErr(certFile, keyFile, clientCAFile string) (grpc.ServerOption, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(clientCAFile)
	if err != nil {
		return nil, err
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("server: no certificates found in client CA file %s", clientCAFile)
	}

	// The internal gateway self-dial must also present a client certificate
	// and trust the same CA, since the gRPC server now requires and
	// verifies client certificates on every connection, including its own
	// gateway's.
	dialCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      clientCAs,
	})

	return tlsOpt{
		certFile:   certFile,
		keyFile:    keyFile,
		dialOpt:    []grpc.DialOption{grpc.WithTransportCredentials(dialCreds)},
		clientAuth: tls.RequireAndVerifyClientCert,
		clientCAs:  clientCAs,
		grpcCreds: credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAs,
		}),
	}, nil
}

// Address serves both gRPC and HTTP traffic on the same address.
//
// This is the simplest option and works well for local development and small
// deployments. If you need separate network policies, independent listeners, or
// fewer grpc-gateway limitations, use SeparateAddresses instead.
func Address(addr string) grpc.ServerOption {
	return addrOpt{
		httpAddr: addr,
		grpcAddr: addr,
	}
}

// SeparateAddresses serves gRPC and HTTP traffic on different addresses.
//
// Prefer this in production when you want independent ports for direct gRPC
// traffic and the HTTP gateway.
func SeparateAddresses(grpcAddr, httpAddr string) grpc.ServerOption {
	return addrOpt{
		httpAddr: httpAddr,
		grpcAddr: grpcAddr,
	}
}

// APIPrefix sets the URL prefix used when mounting grpc-gateway handlers.
//
// For example, passing `/api` mounts generated HTTP handlers under `/api`.
func APIPrefix(prefix string) grpc.ServerOption {
	return apiPrefixOpt{
		prefix: prefix,
	}
}

// Handler registers an additional HTTP handler under pathPrefix.
//
// Use this for health checks, metrics, or custom HTTP endpoints that should be
// served alongside grpc-gateway routes.
func Handler(pathPrefix string, h http.Handler) grpc.ServerOption {
	return handlerOpt{
		prefix: pathPrefix,
		h:      h,
	}
}

// NotFoundHandler sets the handler used for unmatched HTTP routes.
//
// The handler is also used by grpc-gateway routing errors that map to HTTP 404.
func NotFoundHandler(h http.Handler) grpc.ServerOption {
	return notFoundHandlerOpt{
		h: h,
	}
}

// Middlewares applies HTTP middleware to all requests handled by the embedded
// HTTP server, including grpc-gateway routes and custom handlers.
func Middlewares(mdws ...middleware) grpc.ServerOption {
	return mdwOpt{
		mdws: mdws,
	}
}

// Listener uses the same net.Listener for both gRPC and HTTP traffic.
//
// This is the listener equivalent of Address.
func Listener(lis net.Listener) grpc.ServerOption {
	return lisOpt{
		lis:     lis,
		httpLis: lis,
	}
}

// SeparateListeners uses different listeners for gRPC and HTTP traffic.
//
// This is the listener equivalent of SeparateAddresses.
func SeparateListeners(grpcLis, httpLis net.Listener) grpc.ServerOption {
	return lisOpt{
		lis:     grpcLis,
		httpLis: httpLis,
	}
}

// ShutdownTimeout sets how long shutdown waits for in-flight work to finish.
//
// A negative duration keeps the current default behavior. A zero duration asks
// the server to shut down immediately without waiting.
func ShutdownTimeout(d time.Duration) grpc.ServerOption {
	return shutdownTimeout{
		timeout: d,
	}
}

// AutoMaxProcs enables container-aware GOMAXPROCS for this server's process.
//
// It sets GOMAXPROCS to match the CPU quota visible to the process (for
// example, a Kubernetes pod's CPU limit) instead of the host's full CPU
// count, and keeps it updated if the quota changes. This removes the need
// for a separate automaxprocs-style dependency.
//
// It overrides any GOMAXPROCS value set via the GOMAXPROCS environment
// variable or a prior runtime.GOMAXPROCS call, so only enable it when you
// want this package to own that decision.
func AutoMaxProcs() grpc.ServerOption {
	return autoMaxProcsOpt{}
}

// Keepalive configures gRPC server-side keepalive ping parameters and
// enforcement policy. Use this (or KeepaliveDefaults) so long-lived
// connections behind L4 load balancers and NAT gateways aren't silently
// dropped for inactivity.
func Keepalive(params keepalive.ServerParameters, policy keepalive.EnforcementPolicy) grpc.ServerOption {
	return keepaliveOpt{params: params, policy: policy}
}

// KeepaliveDefaults applies production-sensible keepalive settings: ping an
// idle connection every 60s (20s timeout to react), and allow pings even
// without active streams so a fully-idle connection is still probed rather
// than silently dropped by an intermediary.
func KeepaliveDefaults() grpc.ServerOption {
	return Keepalive(
		keepalive.ServerParameters{
			Time:    60 * time.Second,
			Timeout: 20 * time.Second,
		},
		keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		},
	)
}

// WithIncomingHeaderMatcher returns a grpc-gateway option that forwards the
// provided HTTP header names to gRPC metadata.
//
// Header names are canonicalized before matching. The default passthrough
// headers for this package are always included.
func WithIncomingHeaderMatcher(keys []string) runtime.ServeMuxOption {
	merged := append(keys, defaultGWPassthroughHeaders...)
	slices.Sort(merged)
	merged = slices.Compact(merged)
	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, k := range merged {
			if k == canonicalKey || textproto.CanonicalMIMEHeaderKey(k) == canonicalKey {
				return k, true
			}
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}

// WithIncomingHeaderPrefixMatcher returns a grpc-gateway option that forwards
// HTTP headers whose canonicalized names match one of the provided prefixes.
//
// Default passthrough headers are still forwarded even when no prefix matches.
func WithIncomingHeaderPrefixMatcher(prefixes []string) runtime.ServeMuxOption {
	canonicalPrefixes := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		if prefix == "" {
			continue
		}
		canonicalPrefixes = append(canonicalPrefixes, textproto.CanonicalMIMEHeaderKey(prefix))
	}
	slices.Sort(canonicalPrefixes)
	canonicalPrefixes = slices.Compact(canonicalPrefixes)

	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, prefix := range canonicalPrefixes {
			if strings.HasPrefix(canonicalKey, prefix) {
				return canonicalKey, true
			}
		}
		for _, k := range defaultGWPassthroughHeaders {
			if textproto.CanonicalMIMEHeaderKey(k) == canonicalKey {
				return canonicalKey, true
			}
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}
