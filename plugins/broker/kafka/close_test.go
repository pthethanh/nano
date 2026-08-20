package kafka_test

import (
	"context"
	"testing"

	"github.com/pthethanh/nano/plugins/broker/kafka"
)

func TestClose_WithoutOpenDoesNotPanic(t *testing.T) {
	b := kafka.New[string]()
	if err := b.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
