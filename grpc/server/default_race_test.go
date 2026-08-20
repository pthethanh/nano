package server_test

import (
	"sync"
	"testing"

	"github.com/pthethanh/nano/grpc/server"
)

// TestDefault_ConcurrentSetDefaultAndDefaultNeverLosesAWrite is a stress
// test for the fix to a check-then-act race between Default()'s lazy
// construction and a concurrent SetDefault(): the old implementation used
// sync.Once around a Load-then-Store sequence, so a concurrent SetDefault
// landing inside that window could be silently overwritten by the
// lazily-constructed default. That race is one-shot per process (sync.Once
// only ever fires once) and depends on scheduler timing, so it can't be
// reliably forced from a black-box test without flakiness; this test
// instead confirms the fixed behavior holds under concurrent load and is
// race-detector clean.
func TestDefault_ConcurrentSetDefaultAndDefaultNeverLosesAWrite(t *testing.T) {
	old := server.Default()
	t.Cleanup(func() {
		server.SetDefault(old)
	})

	custom := server.New(server.Address("127.0.0.1:0"))

	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			server.SetDefault(custom)
		}()
	}
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = server.Default()
		}()
	}
	close(start)
	wg.Wait()

	if got := server.Default(); got != custom {
		t.Errorf("Default() = %p, want the last SetDefault() value %p", got, custom)
	}
}
