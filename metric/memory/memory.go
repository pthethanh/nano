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
		prefix string
	}
)

func New() *Reporter {
	return &Reporter{}
}

func (r *Reporter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	promhttp.Handler().ServeHTTP(w, req)
}

func (r *Reporter) Counter(name string, labels ...string) metric.Counter {
	return newCounter(name, labels...)
}

func (r *Reporter) Gauge(name string, labels ...string) metric.Gauge {
	return newGauge(name, labels...)
}

func (r *Reporter) Histogram(name string, buckets []float64, labels ...string) metric.Histogram {
	return newHistogram(name, buckets, labels...)
}

func (r *Reporter) Summary(name string, obj map[float64]float64, age time.Duration, labels ...string) metric.Summary {
	return newSummary(name, obj, age, labels...)
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
