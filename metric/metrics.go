package metric

import "time"

type (
	// Reporter provides metric instruments.
	Reporter interface {
		// Counter returns a counter metric.
		Counter(name string, labels ...string) Counter
		// Gauge returns a gauge metric.
		Gauge(name string, labels ...string) Gauge
		// Histogram returns a histogram metric.
		Histogram(name string, buckets []float64, labels ...string) Histogram
		// Summary returns a summary metric.
		Summary(name string, obj map[float64]float64, age time.Duration, labels ...string) Summary
		// Named returns a reporter with the given name.
		Named(name string) Reporter
	}

	// Counter is a metric for counting events.
	Counter interface {
		// With returns a counter with tags.
		With(tags ...string) Counter
		// Add increments the counter by delta.
		Add(delta float64)
	}

	// Gauge is a metric for tracking a value.
	Gauge interface {
		// With returns a gauge with tags.
		With(tags ...string) Gauge
		// Set sets the gauge to value.
		Set(value float64)
		// Add increments the gauge by delta.
		Add(delta float64)
	}

	// Histogram is a metric for recording distributions.
	Histogram interface {
		// With returns a histogram with tags.
		With(tags ...string) Histogram
		// Record adds a value to the histogram.
		Record(value float64)
	}

	// Summary is a metric for tracking quantiles.
	Summary interface {
		// With returns a summary with tags.
		With(tags ...string) Summary
		// Record adds a value to the summary.
		Record(value float64)
	}
)
