package metric

type (
	Reporter interface {
		Counter(name string, labels ...string) Counter
		Gauge(name string, labels ...string) Gauge
		Histogram(name string, buckets []float64, labels ...string) Histogram

		Named(name string) Reporter
	}

	Counter interface {
		With(tags ...string) Counter
		Add(delta float64)
	}

	Gauge interface {
		With(tags ...string) Gauge
		Set(value float64)
		Add(delta float64)
	}

	Histogram interface {
		With(tags ...string) Histogram
		Record(value float64)
	}
)
