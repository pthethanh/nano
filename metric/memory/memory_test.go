package memory_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pthethanh/nano/metric/memory"
)

func BenchmarkRegister(b *testing.B) {
	metrics := memory.New()
	for b.Loop() {
		metrics.Counter("test", "method").With("method", "hello").Add(1)
	}
}

func TestCounter(t *testing.T) {
	metrics := memory.New()
	counter := metrics.Counter("test_counter", "label")
	counter.With("label", "value").Add(5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	metrics.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body), `test_counter{label="value"} 5`) {
		t.Errorf("expected counter value in metrics output, got: %s", string(body))
	}
}

func TestGauge(t *testing.T) {
	metrics := memory.New()
	gauge := metrics.Gauge("test_gauge", "label")
	gauge.With("label", "value").Set(42)
	gauge.With("label", "value").Add(8)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	metrics.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body), `test_gauge{label="value"} 50`) {
		t.Errorf("expected gauge value in metrics output, got: %s", string(body))
	}
}

func TestHistogram(t *testing.T) {
	metrics := memory.New()
	hist := metrics.Histogram("test_hist", []float64{0, 1, 2}, "label")
	hist.With("label", "value").Record(1.5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	metrics.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body), `test_hist_bucket{label="value",le="2"}`) {
		t.Errorf("expected histogram bucket in metrics output, got: %s", string(body))
	}
}

func TestSummary(t *testing.T) {
	metrics := memory.New()
	summary := metrics.Summary("test_summary", map[float64]float64{0.5: 0.05}, 0, "label")
	summary.With("label", "value").Record(2.5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	metrics.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body), `test_summary{label="value",quantile="0.5"}`) {
		t.Errorf("expected summary quantile in metrics output, got: %s", string(body))
	}
}
