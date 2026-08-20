package health

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	// Server is a simple implementation of Server.
	Server struct {
		*health.Server
		cancelCtx context.Context
		cancel    context.CancelFunc
		apiPrefix string

		mu       sync.Mutex
		checkers map[string]context.CancelFunc
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

	// ServerOption is a function that configures the health server.
	ServerOption func(*Server)

	// Service defines a service to be monitored by the health server.
	Service struct {
		Name     string        `json:"name"`
		Delay    time.Duration `json:"delay"`
		Interval time.Duration `json:"interval"`
		Timeout  time.Duration `json:"timeout"`
		Checker  Checker       `json:"-"`
	}
)

var (
	// NoDelay is a constant for no initial delay.
	NoDelay time.Duration = 0
)

// CheckHealth implements Checker interface.
func (c CheckFunc) CheckHealth(ctx context.Context) error {
	return c(ctx)
}

// APIPrefix sets the API prefix for HTTP endpoints.
func APIPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.apiPrefix = prefix
	}
}

// NewServer creates a new health server.
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

// Add adds a health check for a service with intervals and checker.
// The service name is used to identify the service in health checks.
// The delay is the initial delay before the first health check.
// The interval is the time between subsequent health checks.
// The timeout is the maximum time to wait for a health check to complete.
func (s *Server) Add(srv Service) {
	ctx, cancel := context.WithCancel(s.cancelCtx)

	s.mu.Lock()
	if s.checkers == nil {
		s.checkers = make(map[string]context.CancelFunc)
	}
	if prev, ok := s.checkers[srv.Name]; ok {
		prev() // stop any previous checker registered under the same name
	}
	s.checkers[srv.Name] = cancel
	s.mu.Unlock()

	go func() {
		t := srv.Delay
		for {
			select {
			case <-time.After(t):
				s.checkAndUpdate(srv.Name, srv.Timeout, srv.Checker)
			case <-ctx.Done():
				return
			}
			t = srv.Interval
		}
	}()
}

// Remove stops the periodic checker started by Add for name and marks its
// status as UNKNOWN, without affecting any other service or the server as a
// whole. It is a no-op if name was never added (or was already removed).
func (s *Server) Remove(name string) {
	s.mu.Lock()
	cancel, ok := s.checkers[name]
	if ok {
		delete(s.checkers, name)
	}
	s.mu.Unlock()
	if !ok {
		return
	}
	cancel()
	s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_UNKNOWN)
}

// Register registers the health server with a gRPC server.
func (s *Server) Register(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, s)
}

// HTTPHandler returns HTTP handlers for health check endpoints.
func (s *Server) HTTPHandler() (pathPrefix string, h http.Handler) {
	router := http.NewServeMux()
	router.HandleFunc(path.Join(s.apiPrefix, "/check"), s.checkFunc)
	router.HandleFunc(path.Join(s.apiPrefix, "/list"), s.listFunc)
	return s.apiPrefix, router
}

// checkFunc implements http.Handler interface.
// It returns the health status of the service in JSON format, with the HTTP
// status code reflecting the result: 200 when SERVING, 503 otherwise. This
// lets infrastructure that only inspects the status code (many load balancer
// and Kubernetes HTTP probes do) still detect an unhealthy service.
func (s *Server) checkFunc(w http.ResponseWriter, r *http.Request) {
	rs, err := s.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{
		Service: r.URL.Query().Get("service"),
	})
	if err != nil {
		rs = &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}
	}
	writeHealthJSON(w, rs, rs.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING)
}

// listFunc implements http.Handler interface. The response is considered
// unavailable (HTTP 503) if the list call itself fails, or if any listed
// service is not SERVING.
func (s *Server) listFunc(w http.ResponseWriter, r *http.Request) {
	rs, err := s.List(r.Context(), &grpc_health_v1.HealthListRequest{})
	if err != nil {
		writeHealthJSON(w, &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, false)
		return
	}
	ok := true
	for _, status := range rs.GetStatuses() {
		if status.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			ok = false
			break
		}
	}
	writeHealthJSON(w, rs, ok)
}

// writeHealthJSON marshals a proto health message using proto3 JSON (so enum
// values render as their name, e.g. "SERVING", consistent with grpc-gateway
// elsewhere in this framework) and sets the HTTP status code from ok.
func writeHealthJSON(w http.ResponseWriter, m proto.Message, ok bool) {
	b, err := protojson.Marshal(m)
	if err != nil {
		b = fmt.Appendf(nil, `{"status":%d}`, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		ok = false
	}
	if ok {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Write(b)
}

func (s *Server) checkAndUpdate(name string, timeout time.Duration, check Checker) {
	if check == nil {
		s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		s.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rs := make(chan error, 1)
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
		for serviceName, service := range list.Statuses {
			if serviceName == "" {
				continue
			}
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

// Close close underlying resources
func (s *Server) Close() error {
	s.cancel()
	s.Server.Shutdown()
	return nil
}
