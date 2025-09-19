package server

import (
	"net"
	"net/http"
	"net/textproto"
	"slices"
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

// Logger sets a custom logger for server logging.
func Logger(logger logger) grpc.ServerOption {
	return loggerOpt{
		logger: logger,
	}
}

// OnShutdown sets a function to call before server shutdown.
func OnShutdown(f func()) grpc.ServerOption {
	return onShutdownOpt{
		f: f,
	}
}

// GateWayOpts adds options for the gRPC gateway.
func GateWayOpts(opts ...runtime.ServeMuxOption) grpc.ServerOption {
	return gwOpt{
		opts: opts,
	}
}

// Timeout sets read and write timeouts for the internal HTTP server.
func Timeout(read, write time.Duration) grpc.ServerOption {
	return timeoutOpt{
		read:  read,
		write: write,
	}
}

// TLS enables TLS using the provided cert and key files.
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

// Address sets the server address for both gRPC and HTTP on the same port.
// Note: Serving both gRPC and HTTP on a single port has limitations—features like
// service config, retry, and advanced gRPC connection handling may not be fully supported.
// To avoid these limitations, use SeparateAddresses to expose gRPC and HTTP on
// different ports.
func Address(addr string) grpc.ServerOption {
	return addrOpt{
		httpAddr: addr,
		grpcAddr: addr,
	}
}

// SeparateAddresses sets separate addresses for gRPC and HTTP servers
// in case you want to expose them on different ports.
func SeparateAddresses(grpcAddr, httpAddr string) grpc.ServerOption {
	return addrOpt{
		httpAddr: httpAddr,
		grpcAddr: grpcAddr,
	}
}

// APIPrefix sets the API prefix for the gRPC gateway.
func APIPrefix(prefix string) grpc.ServerOption {
	return apiPrefixOpt{
		prefix: prefix,
	}
}

// Handler registers an additional HTTP handler with the given prefix.
func Handler(pathPrefix string, h http.Handler) grpc.ServerOption {
	return handlerOpt{
		prefix: pathPrefix,
		h:      h,
	}
}

// NotFoundHandler sets a custom handler for 404 responses.
func NotFoundHandler(h http.Handler) grpc.ServerOption {
	return notFoundHandlerOpt{
		h: h,
	}
}

// Middlewares applies middleware to all HTTP requests.
func Middlewares(mdws ...middleware) grpc.ServerOption {
	return mdwOpt{
		mdws: mdws,
	}
}

// Listener sets a custom net.Listener for the server.
func Listener(lis net.Listener) grpc.ServerOption {
	return lisOpt{
		lis:     lis,
		httpLis: lis,
	}
}

// SeparateListeners sets separate listeners for gRPC and HTTP servers.
func SeparateListeners(grpcLis, httpLis net.Listener) grpc.ServerOption {
	return lisOpt{
		lis:     grpcLis,
		httpLis: httpLis,
	}
}

// ShutdownTimeout sets the shutdown timeout duration.
func ShutdownTimeout(d time.Duration) grpc.ServerOption {
	return shutdownTimeout{
		timeout: d,
	}
}

// WithIncomingHeaderMatcher customizes which headers are forwarded from HTTP to gRPC.
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
