package memory

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pthethanh/nano/metric"
)

type histogram struct {
	lbvl *lbvl
	hv   *prometheus.HistogramVec
}

func newHistogram(name string, bucket []float64, labels ...string) *histogram {
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Buckets: bucket,
	}, labels)
	prometheus.MustRegister(hv)
	return &histogram{
		lbvl: &lbvl{},
		hv:   hv,
	}
}

func (h *histogram) With(tags ...string) metric.Histogram {
	if len(tags)%2 != 0 {
		panic("With required a key/value pair")
	}
	return &histogram{
		lbvl: h.lbvl.With(tags...),
		hv:   h.hv,
	}
}

func (h *histogram) Record(value float64) {
	h.hv.With(h.lbvl.m).Observe(value)
}
