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
		// grpc
		addr    string
		lis     net.Listener
		grpcSrv *grpc.Server
		grpcGw  *runtime.ServeMux

		// http
		httpAddr string
		httpLis  net.Listener
		httpSrv  *http.Server
		http2Srv *http2.Server

		router *mux.Router

		// options
		logger          logger
		dialOpts        []grpc.DialOption
		serverOpts      []grpc.ServerOption
		gwOpts          []runtime.ServeMuxOption
		secure          bool
		readTimeout     time.Duration
		writeTimeout    time.Duration
		shutdownTimeout time.Duration
		tlsCertFile     string
		tlsKeyFile      string
		onShutdown      func()
		apiPathPrefix   string
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
		RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error
	}

	httpHandler interface {
		HTTPHandler() (pathPrefix string, h http.Handler)
	}
)

var (
	DefaultAddress = ":8000"
)

// New creates a new gRPC server.
func New(opts ...grpc.ServerOption) *Server {
	srv := &Server{
		addr:            DefaultAddress,
		logger:          slog.Default(),
		router:          mux.NewRouter(),
		apiPathPrefix:   "/",
		shutdownTimeout: -1,
	}
	srv.apply(opts...)
	return srv
}

func (srv *Server) httpServer() *http.Server {
	if srv.httpSrv == nil {
		srv.httpSrv = &http.Server{
			Addr:         srv.addr,
			Handler:      srv.handler(),
			ReadTimeout:  srv.readTimeout,
			WriteTimeout: srv.writeTimeout,
		}
	}
	return srv.httpSrv
}

func (srv *Server) grpcGWServer() *runtime.ServeMux {
	if srv.grpcGw == nil {
		srv.dialOpts = append(srv.dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		srv.gwOpts = append(srv.gwOpts, WithIncomingHeaderMatcher(defaultGWPassthroughHeaders))
		srv.grpcGw = runtime.NewServeMux(srv.gwOpts...)
	}
	return srv.grpcGw
}

func (srv *Server) grpcServer() *grpc.Server {
	if srv.grpcSrv == nil {
		srv.grpcSrv = grpc.NewServer(srv.serverOpts...)
	}
	return srv.grpcSrv
}

// ListenAndServe starts the server and serves the provided services with graceful shutdown.
func (srv *Server) ListenAndServe(ctx context.Context, services ...any) error {
	if err := srv.registerServices(ctx, services...); err != nil {
		return err
	}
	if err := srv.listenAndServe(ctx); err != nil {
		return err
	}
	return nil
}

// RegisterService registers a gRPC service.
func (srv *Server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	srv.grpcServer().RegisterService(desc, impl)
	srv.logger.Log(context.TODO(), slog.LevelInfo, "registered service successfully", "name", getTypeName(impl))
}

// ServeMux returns the internal gRPC-Gateway multiplexer.
func (srv *Server) ServeMux() *runtime.ServeMux {
	return srv.grpcGWServer()
}

// DialOpts returns dial options for connecting to the server.
func (srv *Server) DialOpts() []grpc.DialOption {
	return srv.dialOpts
}

// Address returns the grpc server address.
func (srv *Server) Address() string {
	return srv.addr
}

// Addresses returns the http server address.
func (srv *Server) HTTPAddress() string {
	if srv.secure {
		return "https://" + srv.httpAddr
	}
	return "http://" + srv.httpAddr
}

func (srv *Server) apply(opts ...grpc.ServerOption) {
	for _, opt := range opts {
		if customOpt, ok := opt.(customServerOption); ok {
			srv.applyCustomOption(customOpt)
		} else {
			srv.serverOpts = append(srv.serverOpts, opt)
		}
	}
}

func (srv *Server) applyCustomOption(opt customServerOption) {
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
		srv.httpLis = opt.httpLis
		srv.httpAddr = srv.httpLis.Addr().String()
	case addrOpt:
		srv.addr = opt.grpcAddr
		srv.lis = nil
		srv.httpAddr = opt.httpAddr
		srv.httpLis = nil
	case handlerOpt:
		srv.router.PathPrefix(opt.prefix).Handler(opt.h)
	case mdwOpt:
		for _, mdw := range opt.mdws {
			srv.router.Use(mdw)
		}
	case shutdownTimeout:
		srv.shutdownTimeout = opt.timeout
	default:
		// should not happen
		srv.logger.Log(context.Background(), slog.LevelWarn, "unknown custom server option", "type", fmt.Sprintf("%T", opt))
	}
}

func (srv *Server) listenAndServe(ctx context.Context) error {
	if srv.onShutdown != nil {
		defer srv.onShutdown()
	}
	if srv.lis == nil {
		lis, err := net.Listen("tcp", srv.addr)
		if err != nil {
			return err
		}
		srv.lis = lis
	}
	useSeparateAddresses := !srv.useSingleAddress()
	if useSeparateAddresses && srv.httpLis == nil {
		lis, err := net.Listen("tcp", srv.httpAddr)
		if err != nil {
			return err
		}
		srv.httpLis = lis
	}
	errs := make(chan error, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// start servers
	switch {
	case useSeparateAddresses:
		// start grpc server
		go func() {
			errs <- srv.grpcServer().Serve(srv.lis)
		}()
		srv.logger.Log(ctx, slog.LevelInfo, "gRPC server started", "grpc_address", srv.addr)
		// start http server
		go func() {
			if srv.secure {
				errs <- srv.httpServer().ServeTLS(srv.httpLis, srv.tlsCertFile, srv.tlsKeyFile)
				return
			}
			errs <- srv.httpServer().Serve(srv.httpLis)
		}()
		srv.logger.Log(ctx, slog.LevelInfo, "HTTP server started", "http_address", srv.httpAddr)
	default:
		go func() {
			if srv.secure {
				errs <- srv.httpServer().ServeTLS(srv.lis, srv.tlsCertFile, srv.tlsKeyFile)
				return
			}
			errs <- srv.httpServer().Serve(srv.lis)
		}()
		srv.logger.Log(ctx, slog.LevelInfo, "HTTP & GRPC server started", slog.String("address", srv.addr))
	}

	// handle gracefully shutdown
	select {
	case <-ctx.Done():
		if err := srv.shutdown(ctx); err != nil {
			return err
		}
		return context.Cause(ctx)
	case err := <-errs:
		return err
	case s := <-sigs:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			if err := srv.shutdown(ctx); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (srv *Server) shutdown(ctx context.Context) error {
	srv.logger.Log(ctx, slog.LevelInfo, "server is shutting down")
	newCtx := ctx
	if srv.shutdownTimeout >= 0 {
		ctx, cancel := context.WithTimeout(ctx, srv.shutdownTimeout)
		newCtx = ctx
		defer cancel()
	}
	if srv.grpcSrv != nil {
		srv.grpcServer().GracefulStop()
	}
	return srv.httpServer().Shutdown(newCtx)
}

func (srv *Server) registerServices(ctx context.Context, services ...any) error {
	for _, s := range services {
		serviceDetails := []string{}
		valid := false
		if h, ok := s.(service); ok {
			h.Register(srv.grpcServer())
			valid = true
			serviceDetails = append(serviceDetails, "gRPC")
		} else if h, ok := s.(serviceDescriptor); ok {
			srv.grpcServer().RegisterService(h.ServiceDesc(), h)
			valid = true
			serviceDetails = append(serviceDetails, "gRPC")
		}
		if h, ok := s.(grpcEndpoint); ok {
			if err := h.RegisterWithEndpoint(ctx, srv.grpcGWServer(), srv.addr, srv.dialOpts); err != nil {
				return err
			}
			valid = true
			serviceDetails = append(serviceDetails, "gRPC Gateway")
		}
		if h, ok := s.(httpHandler); ok {
			prefix, h := h.HTTPHandler()
			srv.router.PathPrefix(prefix).Handler(h)
			valid = true
			serviceDetails = append(serviceDetails, "HTTP "+prefix)
		} else if h, ok := s.(http.Handler); ok {
			srv.router.Handle("/", h)
			valid = true
			serviceDetails = append(serviceDetails, "HTTP /")
		}
		if !valid {
			return ErrUnknownServiceType
		}
		srv.logger.Log(ctx, slog.LevelInfo, "registered service successfully", "name", getTypeName(s), "details", serviceDetails)
	}
	srv.router.PathPrefix(srv.apiPathPrefix).Handler(srv.grpcGWServer())
	return nil
}

func (srv *Server) handler() http.Handler {
	if srv.useSingleAddress() {
		return srv.singleAddressHandler()
	}
	return srv.router
}

// singleAddressHandler returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func (srv *Server) singleAddressHandler() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			srv.grpcServer().ServeHTTP(w, r)
			return
		}
		srv.router.ServeHTTP(w, r)
	})
	if srv.secure {
		return h
	}
	// Work-around in case TLS is disabled.
	// See: https://github.com/grpc/grpc-go/issues/555
	srv.http2Srv = &http2.Server{}
	return h2c.NewHandler(h, srv.http2Srv)
}

func (srv *Server) useSingleAddress() bool {
	_, port1, err1 := net.SplitHostPort(srv.addr)
	_, port2, err2 := net.SplitHostPort(srv.httpAddr)
	return err1 == nil && err2 == nil && port1 == port2
}
