package nats_test

import (
	"context"
	"testing"

	"github.com/pthethanh/nano/plugins/broker/nats"
)

func TestClose_WithoutOpenDoesNotPanic(t *testing.T) {
	b := nats.New[string]()
	if err := b.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
