package memory

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pthethanh/nano/metric"
)

type gauge struct {
	lbvl *lbvl
	cv   *prometheus.GaugeVec
}

func newGauge(name string, labels ...string) *gauge {
	cv := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
	}, labels)
	prometheus.MustRegister(cv)
	return &gauge{
		lbvl: &lbvl{},
		cv:   cv,
	}
}

func (c *gauge) Add(delta float64) {
	c.cv.With(c.lbvl.m).Add(delta)
}

func (c *gauge) Set(value float64) {
	c.cv.With(c.lbvl.m).Set(value)
}

func (c *gauge) With(labelValues ...string) metric.Gauge {
	if len(labelValues)%2 != 0 {
		panic("With required a key/value pair")
	}
	cc := &gauge{
		lbvl: c.lbvl.With(labelValues...),
		cv:   c.cv,
	}
	return cc
}
