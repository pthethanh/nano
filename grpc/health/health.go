package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type (
	// Server is a simple implementation of Server.
	Server struct {
		*health.Server
		cancelCtx context.Context
		cancel    context.CancelFunc
		apiPrefix string
	}

	// CheckFunc is quick way to define a health checker.
	CheckFunc func(context.Context) error

	// Checker provide functionality for checking health of a service.
	Checker interface {
		// CheckHealth establish health check to the target service.
		// Return error if target service cannot be reached
		// or not working properly.
		CheckHealth(ctx context.Context) error
	}

	ServerOption func(*Server)
)

var (
	NoDelay time.Duration = 0
)

// CheckHealth implements Checker interface.
func (c CheckFunc) CheckHealth(ctx context.Context) error {
	return c(ctx)
}

func APIPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.apiPrefix = prefix
	}
}

// NewServer return new gRPC health server.
func NewServer(opts ...ServerOption) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{
		Server:    health.NewServer(),
		cancelCtx: ctx,
		cancel:    cancel,
		apiPrefix: "/api/v1/health",
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// AddService add new health check services
func (s *Server) AddService(service string, delay, interval, timeout time.Duration, checker Checker) {
	go func() {
		t := delay
		for {
			select {
			case <-time.After(t):
				s.checkAndUpdate(service, timeout, checker)
			case <-s.cancelCtx.Done():
				return
			}
			t = interval
		}
	}()
}

// Register implements health.Server.
func (s *Server) Register(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, s)
}

func (s *Server) HTTPHandler() (pathPrefix string, h http.Handler) {
	router := http.NewServeMux()
	router.HandleFunc(path.Join(s.apiPrefix, "/check"), s.checkFunc)
	router.HandleFunc(path.Join(s.apiPrefix, "/list"), s.listFunc)
	return s.apiPrefix, router
}

// checkFunc implements http.Handler interface.
// It returns the health status of the service in JSON format.
func (s *Server) checkFunc(w http.ResponseWriter, r *http.Request) {
	rs, err := s.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{
		Service: r.URL.Query().Get("service"),
	})
	if err != nil {
		b, err := json.Marshal(grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		})
		if err != nil {
			b = fmt.Appendf(nil, `{"status":%d}`, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
	b, err := json.Marshal(rs)
	if err != nil {
		b = fmt.Appendf(nil, `{"status":%d}`, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Server) listFunc(w http.ResponseWriter, r *http.Request) {
	rs, err := s.List(r.Context(), &grpc_health_v1.HealthListRequest{})
	if err != nil {
		b, err := json.Marshal(grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		})
		if err != nil {
			b = fmt.Appendf(nil, `{"":%d}`, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
	b, err := json.Marshal(rs)
	if err != nil {
		b = fmt.Appendf(nil, `{"":%d}`, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Server) checkAndUpdate(name string, timeout time.Duration, check Checker) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rs := make(chan error)
	go func() {
		rs <- check.CheckHealth(timeoutCtx)
	}()
	select {
	case err := <-rs:
		if err != nil {
			s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
			s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
			return
		}
		s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)

		// check and update overall status
		list, err := s.List(context.Background(), &grpc_health_v1.HealthListRequest{})
		if err != nil {
			// failed to check, leave it as-is
			return
		}
		for _, service := range list.Statuses {
			if service.Status != grpc_health_v1.HealthCheckResponse_SERVING {
				s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				return
			}
		}
		s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		return
	case <-timeoutCtx.Done():
		s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
		s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_UNKNOWN)
		return
	case <-s.cancelCtx.Done():
		s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return
	}
}

func (s *Server) Close() error {
	s.cancel()
	s.Server.Shutdown()
	return nil
}
