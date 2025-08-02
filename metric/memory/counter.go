package memory

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pthethanh/nano/metric"
)

type counter struct {
	lbvl *lbvl
	cv   *prometheus.CounterVec
}

func newCounter(name string, labels ...string) *counter {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
	}, labels)
	prometheus.MustRegister(cv)
	return &counter{
		lbvl: &lbvl{},
		cv:   cv,
	}
}

func (c *counter) Add(delta float64) {
	c.cv.With(c.lbvl.m).Add(delta)
}

func (c *counter) With(labelValues ...string) metric.Counter {
	if len(labelValues)%2 != 0 {
		panic("With required a key/value pair")
	}
	cc := &counter{
		lbvl: c.lbvl.With(labelValues...),
		cv:   c.cv,
	}
	return cc
}
