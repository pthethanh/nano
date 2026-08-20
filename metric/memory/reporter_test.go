package memory_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pthethanh/nano/metric/memory"
)

func TestNamed_DoesNotPanicAndPrefixesMetricName(t *testing.T) {
	m := memory.New()
	sub := m.Named("sub")

	counter := sub.Counter("requests", "label")
	counter.With("label", "v").Add(1)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	m.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)

	if !strings.Contains(string(body), `sub_requests{label="v"} 1`) {
		t.Errorf("Named(\"sub\").Counter(\"requests\") should register as sub_requests, got output:\n%s", body)
	}
}

func TestNamed_NestedPrefixesCompound(t *testing.T) {
	m := memory.New()
	sub := m.Named("a").Named("b")

	sub.Counter("hits").Add(1)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	m.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Result().Body)

	if !strings.Contains(string(body), "a_b_hits 1") {
		t.Errorf("nested Named(\"a\").Named(\"b\").Counter(\"hits\") should register as a_b_hits, got output:\n%s", body)
	}
}

func TestTwoReporters_SameMetricName_DoNotPanic(t *testing.T) {
	m1 := memory.New()
	m2 := memory.New()

	// Two independently constructed Reporters requesting the same metric
	// name must not collide on a shared global registry.
	m1.Counter("shared_name", "label").With("label", "x").Add(1)
	m2.Counter("shared_name", "label").With("label", "x").Add(1)
}
