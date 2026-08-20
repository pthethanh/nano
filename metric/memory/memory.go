// package memory implement in-mem metrics using prometheus lib.
package memory

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pthethanh/nano/metric"
)

type (
	Reporter struct {
		apiPrefix string
		registry  *prometheus.Registry

		prefix     string
		counters   *cache[*counter]
		gauges     *cache[*gauge]
		histograms *cache[*histogram]
		summaries  *cache[*summary]
	}
	ReporterOption func(*Reporter)
)

func APIPrefix(prefix string) ReporterOption {
	return func(r *Reporter) {
		r.apiPrefix = prefix
	}
}

// New creates a Reporter backed by its own prometheus.Registry, so metric
// names only need to be unique within one Reporter (and its Named
// sub-reporters) rather than across every Reporter in the process.
func New(opts ...ReporterOption) *Reporter {
	r := &Reporter{
		apiPrefix:  "/api/v1/metrics",
		registry:   prometheus.NewRegistry(),
		counters:   newCache[*counter](),
		summaries:  newCache[*summary](),
		histograms: newCache[*histogram](),
		gauges:     newCache[*gauge](),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Reporter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{}).ServeHTTP(w, req)
}

func (r *Reporter) HTTPHandler() (string, http.Handler) {
	return r.apiPrefix, http.HandlerFunc(r.ServeHTTP)
}

func (r *Reporter) Counter(name string, labels ...string) metric.Counter {
	name = r.prefixed(name)
	return r.counters.loadOrCreate(name, labels, func() *counter {
		return newCounter(r.registry, name, labels...)
	})
}

func (r *Reporter) Gauge(name string, labels ...string) metric.Gauge {
	name = r.prefixed(name)
	return r.gauges.loadOrCreate(name, labels, func() *gauge {
		return newGauge(r.registry, name, labels...)
	})
}

func (r *Reporter) Histogram(name string, buckets []float64, labels ...string) metric.Histogram {
	name = r.prefixed(name)
	return r.histograms.loadOrCreate(name, labels, func() *histogram {
		return newHistogram(r.registry, name, buckets, labels...)
	})
}

func (r *Reporter) Summary(name string, obj map[float64]float64, age time.Duration, labels ...string) metric.Summary {
	name = r.prefixed(name)
	return r.summaries.loadOrCreate(name, labels, func() *summary {
		return newSummary(r.registry, name, obj, age, labels...)
	})
}

// Named returns a sub-reporter whose metric names are prefixed with name
// (joined by "_"), registered against the same underlying registry so all
// metrics remain visible on the parent Reporter's HTTP endpoint.
func (r *Reporter) Named(name string) metric.Reporter {
	return &Reporter{
		apiPrefix:  r.apiPrefix,
		registry:   r.registry,
		prefix:     r.prefixed(name),
		counters:   r.counters,
		gauges:     r.gauges,
		histograms: r.histograms,
		summaries:  r.summaries,
	}
}

func (r *Reporter) prefixed(name string) string {
	if r.prefix == "" {
		return name
	}
	return r.prefix + "_" + name
}
