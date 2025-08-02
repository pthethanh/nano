package memory

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/pthethanh/nano/metric"
)

var (
	defaultObjective = map[float64]float64{
		0.00: 0.1,
		0.25: 0.01,
		0.50: 0.01,
		0.75: 0.001,
		1.00: 0.0001,
	}
	defaultMaxAge = 10 * time.Minute
)

type summary struct {
	lbvl *lbvl
	hv   *prometheus.SummaryVec
}

func newSummary(name string, objectives map[float64]float64, maxAge time.Duration, labels ...string) *summary {
	obj := objectives
	if len(objectives) == 0 {
		obj = defaultObjective
	}
	age := maxAge
	if age == 0 {
		age = defaultMaxAge
	}
	hv := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       name,
		Objectives: obj,
		MaxAge:     age,
	}, labels)
	prometheus.MustRegister(hv)
	return &summary{
		lbvl: &lbvl{},
		hv:   hv,
	}
}

func (h *summary) With(tags ...string) metric.Summary {
	if len(tags)%2 != 0 {
		panic("With required a key/value pair")
	}
	return &summary{
		lbvl: h.lbvl.With(tags...),
		hv:   h.hv,
	}
}

func (h *summary) Record(value float64) {
	h.hv.With(h.lbvl.m).Observe(value)
}
