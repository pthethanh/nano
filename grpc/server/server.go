package server

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

func New(opts ...grpc.ServerOption) *Server {
	srv := &Server{
		addr:          ":8000",
		logger:        slog.Default(),
		router:        mux.NewRouter(),
		http2Srv:      &http2.Server{},
		apiPathPrefix: "/",
		dialOpts:      []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		gwOpts:        []runtime.ServeMuxOption{headerMatcher([]string{"X-Request-Id", "X-Correlation-ID", "Api-Key"})},
	}
	srv.apply(opts...)
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

func (srv *Server) apply(opts ...grpc.ServerOption) {
	for _, opt := range opts {
		if _, ok := opt.(customServerOption); ok {
			switch opt := opt.(type) {
			case loggerOpt:
				srv.logger = opt.logger
			case onShutdownOpt:
				srv.onShutdown = opt.f
			case gwOpt:
				srv.gwOpts = append(srv.gwOpts, opt.opts...)
			case notFoundHandlerOpt:
				srv.router.NotFoundHandler = opt.h
				srv.gwOpts = append(srv.gwOpts, runtime.WithRoutingErrorHandler(func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, i int) {
					if http.StatusNotFound == i {
						opt.h.ServeHTTP(w, r)
						return
					}
					runtime.DefaultRoutingErrorHandler(ctx, sm, m, w, r, i)
				}))
			case apiPrefixOpt:
				srv.apiPathPrefix = opt.prefix
			case timeoutOpt:
				srv.readTimeout = opt.read
				srv.writeTimeout = opt.write
			case tlsOpt:
				srv.tlsCertFile = opt.certFile
				srv.tlsKeyFile = opt.keyFile
				srv.secure = true
				srv.dialOpts = append(srv.dialOpts, opt.dialOpt...)
			case lisOpt:
				srv.lis = opt.lis
				srv.addr = opt.lis.Addr().String()
			case addrOpt:
				srv.addr = opt.addr
				srv.lis = nil
			case handlerOpt:
				srv.router.PathPrefix(opt.prefix).Handler(opt.h)
			case mdwOpt:
				for _, mdw := range opt.mdws {
					srv.router.Use(mdw)
				}
			}
		} else {
			fmt.Printf("%#v\n", opt)
			srv.serverOpts = append(srv.serverOpts, opt)
		}
	}
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
