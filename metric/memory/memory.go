// package memory implement in-mem metrics using prometheus lib.
package memory

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pthethanh/nano/metric"
)

type (
	Reporter struct {
		prefix     string
		counters   *cache[*counter]
		gauges     *cache[*gauge]
		histograms *cache[*histogram]
		summaries  *cache[*summary]
	}
)

func New() *Reporter {
	return &Reporter{
		counters:   newCache[*counter](),
		summaries:  newCache[*summary](),
		histograms: newCache[*histogram](),
		gauges:     newCache[*gauge](),
	}
}

func (r *Reporter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	promhttp.Handler().ServeHTTP(w, req)
}

func (r *Reporter) Counter(name string, labels ...string) metric.Counter {
	return r.counters.loadOrCreate(name, labels, func() *counter {
		return newCounter(name, labels...)
	})
}

func (r *Reporter) Gauge(name string, labels ...string) metric.Gauge {
	return r.gauges.loadOrCreate(name, labels, func() *gauge {
		return newGauge(name, labels...)
	})
}

func (r *Reporter) Histogram(name string, buckets []float64, labels ...string) metric.Histogram {
	return r.histograms.loadOrCreate(name, labels, func() *histogram {
		return newHistogram(name, buckets, labels...)
	})
}

func (r *Reporter) Summary(name string, obj map[float64]float64, age time.Duration, labels ...string) metric.Summary {
	return r.summaries.loadOrCreate(name, labels, func() *summary {
		return newSummary(name, obj, age, labels...)
	})
}

func (r *Reporter) Named(name string) metric.Reporter {
	newName := name
	if r.prefix != "" {
		newName = r.prefix + "_" + name
	}
	return &Reporter{
		prefix: newName,
	}
}
