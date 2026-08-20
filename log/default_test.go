package log_test

import (
	"testing"

	"github.com/pthethanh/nano/log"
)

func TestSetDefaultNil_DoesNotPermanentlyBreakDefault(t *testing.T) {
	_ = log.Default() // ensure a real default logger has been lazily constructed once

	log.SetDefault(nil)

	got := log.Default()
	if got == nil {
		t.Fatal("Default() returned nil after SetDefault(nil): the default logger is permanently broken, and the next log.Info(...) call will nil-deref")
	}
}
