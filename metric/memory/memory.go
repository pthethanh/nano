// package memory implement in-mem metrics using prometheus lib.
package memory

import (
	"net/http"

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
	if buckets == nil {
		return newHistogram(name, nil, labels...)
	}
	return newHistogram(name, buckets, labels...)
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
