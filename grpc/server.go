package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Server struct {
		addr     string
		logger   logger
		services []any

		// grpc
		lis           net.Listener
		grpcSrv       *grpc.Server
		httpSrv       *http.Server
		http2Srv      *http2.Server
		gw            *runtime.ServeMux
		dialOpts      []grpc.DialOption
		serverOpts    []grpc.ServerOption
		gwOpts        []runtime.ServeMuxOption
		secure        bool
		readTimeout   time.Duration
		writeTimeout  time.Duration
		tlsCertFile   string
		tlsKeyFile    string
		onShutdown    func()
		apiPathPrefix string

		// normal http router
		router *mux.Router
	}

	ServerOption func(srv *Server)

	service interface {
		Register(srv *grpc.Server)
	}

	// serviceDescriptor implements grpc service that expose its service desc.
	serviceDescriptor interface {
		ServiceDesc() *grpc.ServiceDesc
	}

	// grpcEndpoint implement an endpoint registration interface for service to attach their endpoints to gRPC gateway.
	grpcEndpoint interface {
		RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption)
	}

	httpHandler interface {
		HTTPHandler() (pathPrefix string, h http.Handler)
	}
)

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		addr:          ":8000",
		logger:        slog.Default(),
		router:        mux.NewRouter(),
		http2Srv:      &http2.Server{},
		apiPathPrefix: "/",
	}
	for _, apply := range opts {
		apply(srv)
	}
	// apply default options if not defined.
	if len(srv.dialOpts) == 0 {
		srv.dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	if len(srv.gwOpts) == 0 {
		srv.gwOpts = []runtime.ServeMuxOption{
			headerMatcher([]string{"X-Request-Id", "X-Correlation-ID", "Api-Key"}),
		}
	}
	srv.grpcSrv = grpc.NewServer(srv.serverOpts...)
	srv.gw = runtime.NewServeMux(srv.gwOpts...)
	srv.httpSrv = &http.Server{
		Addr:         srv.addr,
		Handler:      srv.handler(),
		ReadTimeout:  srv.readTimeout,
		WriteTimeout: srv.writeTimeout,
	}
	if srv.onShutdown != nil {
		srv.httpSrv.RegisterOnShutdown(srv.onShutdown)
	}
	return srv
}

func (srv *Server) ListenAndServe(ctx context.Context, services ...any) error {
	if err := srv.registerServices(ctx, services...); err != nil {
		return err
	}
	if err := srv.listenAndServe(ctx); err != nil {
		return err
	}
	return nil
}

func (srv *Server) listenAndServe(ctx context.Context) error {
	if srv.lis == nil {
		lis, err := net.Listen("tcp", srv.addr)
		if err != nil {
			return err
		}
		srv.lis = lis
	}
	errs := make(chan error, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		if srv.secure {
			errs <- srv.httpSrv.ServeTLS(srv.lis, srv.tlsCertFile, srv.tlsKeyFile)
			return
		}
		errs <- srv.httpSrv.Serve(srv.lis)
	}()
	srv.logger.Log(ctx, slog.LevelInfo, "server started", slog.String("address", srv.addr))

	// handle gracefully shutdown
	select {
	case <-ctx.Done():
		srv.logger.Log(ctx, slog.LevelInfo, "server is shutting down")
		if err := srv.httpSrv.Shutdown(ctx); err != nil {
			return err
		}
		return context.Cause(ctx)
	case err := <-errs:
		return err
	case s := <-sigs:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			srv.logger.Log(ctx, slog.LevelInfo, "server is shutting down")
			if err := srv.httpSrv.Shutdown(ctx); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (srv *Server) registerServices(ctx context.Context, services ...any) error {
	for _, s := range services {
		valid := false
		if h, ok := s.(service); ok {
			h.Register(srv.grpcSrv)
			valid = true
		} else if h, ok := s.(serviceDescriptor); ok {
			srv.grpcSrv.RegisterService(h.ServiceDesc(), h)
			valid = true
		}
		if h, ok := s.(grpcEndpoint); ok {
			h.RegisterWithEndpoint(ctx, srv.gw, srv.addr, srv.dialOpts)
			valid = true
		}
		if h, ok := s.(httpHandler); ok {
			prefix, h := h.HTTPHandler()
			srv.router.PathPrefix(prefix).Handler(h)
			valid = true
		}
		if h, ok := s.(http.Handler); ok {
			srv.router.Handle("", h)
			valid = true
		}
		if !valid {
			return ErrUnknownServiceType
		}
		srv.logger.Log(ctx, slog.LevelInfo, "registered service successfully", "name", fmt.Sprintf("%T", s))
	}
	srv.router.PathPrefix(srv.apiPathPrefix).Handler(srv.gw)
	return nil
}

// handler returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func (srv *Server) handler() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			srv.grpcSrv.ServeHTTP(w, r)
			return
		}
		srv.router.ServeHTTP(w, r)
	})
	if srv.secure {
		return h
	}
	// Work-around in case TLS is disabled.
	// See: https://github.com/grpc/grpc-go/issues/555
	return h2c.NewHandler(h, srv.http2Srv)
}