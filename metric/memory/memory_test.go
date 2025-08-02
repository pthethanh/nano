package memory_test

import (
	"testing"

	"github.com/pthethanh/nano/metric/memory"
)

var (
	metrics = memory.New()
)

func BenchmarkRegister(b *testing.B) {
	for i := 0; i < b.N; i++ {
		metrics.Counter("test", "method").With("method", "hello").Add(1)
	}
}
