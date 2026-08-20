package validator_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pthethanh/nano/validator"
)

func TestDefault_ConcurrentFirstCallForUnseenTagReturnsSameInstance(t *testing.T) {
	tag := fmt.Sprintf("concurrent-tag-%p", t) // unique per test run, never seen before

	var wg sync.WaitGroup
	results := make([]*validator.Validator, 50)
	start := make(chan struct{})
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i] = validator.Default(tag)
		}(i)
	}
	close(start)
	wg.Wait()

	first := results[0]
	for i, v := range results {
		if v != first {
			t.Fatalf("Default(%q) returned a different instance for goroutine %d (%p) than goroutine 0 (%p): concurrent first-time construction is not atomic", tag, i, v, first)
		}
	}
}
