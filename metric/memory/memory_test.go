package memory_test

import (
	"testing"

	"github.com/pthethanh/nano/metric/memory"
)

func BenchmarkRegister(b *testing.B) {
	metrics := memory.New()
	for b.Loop() {
		metrics.Counter("test", "method").With("method", "hello").Add(1)
	}
}
