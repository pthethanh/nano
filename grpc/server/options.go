package server

import (
	"net"
	"net/http"
	"net/textproto"
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
		addr string
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
		lis net.Listener
	}

	shutdownTimeout struct {
		emptyOpt
		timeout time.Duration
	}
)

// Logger provide alternate logger for server logging
func Logger(logger logger) grpc.ServerOption {
	return loggerOpt{
		logger: logger,
	}
}

// OnShutdown provide custom func to be called before shutting down
func OnShutdown(f func()) grpc.ServerOption {
	return onShutdownOpt{
		f: f,
	}
}

// GateWayOpts provide additional options for api gateway
func GateWayOpts(opts ...runtime.ServeMuxOption) grpc.ServerOption {
	return gwOpt{
		opts: opts,
	}
}

// Timeout set read, write timeout for internal http server.
func Timeout(read, write time.Duration) grpc.ServerOption {
	return timeoutOpt{
		read:  read,
		write: write,
	}
}

// TLS enable secure mode using tls key & cert file.
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

// Address set server address
func Address(addr string) grpc.ServerOption {
	return addrOpt{
		addr: addr,
	}
}

// APIPrefix defines grpc gateway api prefix
func APIPrefix(prefix string) grpc.ServerOption {
	return apiPrefixOpt{
		prefix: prefix,
	}
}

// Handler provide ability to define additional HTTP apis beside GRPC Gateway API.
// All HTTP API with the given prefix will be forwarded to the given handler.
func Handler(pathPrefix string, h http.Handler) grpc.ServerOption {
	return handlerOpt{
		prefix: pathPrefix,
		h:      h,
	}
}

// NotFoundHandler provide alternative not found HTTP handler.
func NotFoundHandler(h http.Handler) grpc.ServerOption {
	return notFoundHandlerOpt{
		h: h,
	}
}

// Middlewares apply the given middleware on all HTTP requests
func Middlewares(mdws ...middleware) grpc.ServerOption {
	return mdwOpt{
		mdws: mdws,
	}
}

// Listener force server to use the given listener.
func Listener(lis net.Listener) grpc.ServerOption {
	return lisOpt{
		lis: lis,
	}
}

// ShutdownTimeout define timeout on shutdown.
func ShutdownTimeout(d time.Duration) grpc.ServerOption {
	return shutdownTimeout{
		timeout: d,
	}
}

func headerMatcher(keys []string) runtime.ServeMuxOption {
	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, k := range keys {
			if k == canonicalKey || textproto.CanonicalMIMEHeaderKey(k) == canonicalKey {
				return k, true
			}
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}
