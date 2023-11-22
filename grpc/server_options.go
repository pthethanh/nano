package grpc

import (
	"context"
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
)

// Logger provide alternate logger for server logging
func Logger(logger logger) ServerOption {
	return func(srv *Server) {
		srv.logger = logger
	}
}

// OnShutdown provide custom func to be called before shutting down
func OnShutdown(f func()) ServerOption {
	return func(srv *Server) {
		srv.onShutdown = f
	}
}

// ServerOpts provides additional grpc server opts for server creation.
func ServerOpts(opts ...grpc.ServerOption) ServerOption {
	return func(srv *Server) {
		srv.serverOpts = opts
	}
}

// GateWayOpts provide additional options for api gateway
func GateWayOpts(opts ...runtime.ServeMuxOption) ServerOption {
	return func(srv *Server) {
		srv.gwOpts = opts
	}
}

// Timeout set read, write timeout for internal http server.
func Timeout(read, write time.Duration) ServerOption {
	return func(srv *Server) {
		srv.readTimeout = read
		srv.writeTimeout = write
	}
}

// TLS enable secure mode using tls key & cert file.
func TLS(certFile, keyFile string) ServerOption {
	return func(srv *Server) {
		srv.tlsCertFile = certFile
		srv.tlsKeyFile = keyFile
		srv.secure = true
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			panic(err)
		}
		srv.dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	}
}

// Address set server address
func Address(addr string) ServerOption {
	return func(srv *Server) {
		srv.addr = addr
	}
}

// APIPrefix defines grpc gateway api prefix
func APIPrefix(prefix string) ServerOption {
	return func(srv *Server) {
		srv.apiPathPrefix = prefix
	}
}

// Handler provide ability to define additional HTTP apis beside GRPC Gateway API
func Handler(pathPrefix string, h http.Handler) ServerOption {
	return func(srv *Server) {
		srv.router.PathPrefix(pathPrefix).Handler(h)
	}
}

// NotFoundHandler provide alternative not found HTTP handler.
func NotFoundHandler(h http.Handler) ServerOption {
	return func(srv *Server) {
		srv.router.NotFoundHandler = h
		srv.gwOpts = append(srv.gwOpts, runtime.WithRoutingErrorHandler(func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, i int) {
			if http.StatusNotFound == i {
				h.ServeHTTP(w, r)
				return
			}
			runtime.DefaultRoutingErrorHandler(ctx, sm, m, w, r, i)
		}))
	}
}

// Middlewares apply the given middleware on all HTTP requests
func Middlewares(mdws ...middleware) ServerOption {
	return func(srv *Server) {
		for _, mdw := range mdws {
			srv.router.Use(mdw)
		}
	}
}

// Listener force server to use the given listener.
func Listener(lis net.Listener) ServerOption {
	return func(srv *Server) {
		srv.lis = lis
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
