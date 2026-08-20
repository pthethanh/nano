package health_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pthethanh/nano/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestHTTPCheck_ReturnsServiceUnavailableWhenNotServing(t *testing.T) {
	s := health.NewServer()
	defer s.Close()
	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	_, handler := s.HTTPHandler()
	req := httptest.NewRequest("GET", "/api/v1/health/check", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 503 {
		t.Errorf("HTTP status = %d, want 503 for a NOT_SERVING status (infrastructure that only reads the status code must be able to detect this)", w.Code)
	}
}

func TestHTTPCheck_ReturnsOKWhenServing(t *testing.T) {
	s := health.NewServer()
	defer s.Close()
	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	_, handler := s.HTTPHandler()
	req := httptest.NewRequest("GET", "/api/v1/health/check", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("HTTP status = %d, want 200 for a SERVING status", w.Code)
	}
}

func TestHTTPCheck_ChecksUnknownServiceReturnsUnavailable(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	_, handler := s.HTTPHandler()
	req := httptest.NewRequest("GET", "/api/v1/health/check?service=does-not-exist", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 503 {
		t.Errorf("HTTP status = %d, want 503 when Check() itself errors (unknown service)", w.Code)
	}
}

func TestHTTPCheck_BodyUsesProtoJSONEnumNames(t *testing.T) {
	s := health.NewServer()
	defer s.Close()
	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	_, handler := s.HTTPHandler()
	req := httptest.NewRequest("GET", "/api/v1/health/check", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body), `"SERVING"`) {
		t.Errorf("body = %s, want proto3-JSON enum name \"SERVING\" (encoding/json on the raw struct emits the integer instead)", body)
	}
}

func TestHTTPList_ReturnsServiceUnavailableWhenAnyServiceNotServing(t *testing.T) {
	s := health.NewServer()
	defer s.Close()
	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.SetServingStatus("degraded-dep", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	_, handler := s.HTTPHandler()
	req := httptest.NewRequest("GET", "/api/v1/health/list", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 503 {
		t.Errorf("HTTP status = %d, want 503 when any listed service is not SERVING", w.Code)
	}
}
