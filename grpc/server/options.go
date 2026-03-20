package server

import (
	"net"
	"net/http"
	"net/textproto"
	"slices"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		certFile string
		keyFile  string
		dialOpt  []grpc.DialOption
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
func TLS(certFile, keyFile string) grpc.ServerOption {
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		panic(err)
	}
	return tlsOpt{
		keyFile:  keyFile,
		certFile: certFile,
		dialOpt:  []grpc.DialOption{grpc.WithTransportCredentials(creds)},
	}
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
