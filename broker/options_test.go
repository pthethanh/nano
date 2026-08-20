package broker_test

import (
	"reflect"
	"testing"

	"github.com/pthethanh/nano/broker"
)

func TestPublishOptions_HeaderAndHeaders(t *testing.T) {
	opts := &broker.PublishOptions{}
	opts.Apply(
		broker.Headers(map[string]string{"a": "1", "b": "2"}),
		broker.Header("c", "3"),
	)

	want := map[string]string{"a": "1", "b": "2", "c": "3"}
	if !reflect.DeepEqual(opts.Headers, want) {
		t.Errorf("Headers = %v, want %v", opts.Headers, want)
	}
}

func TestPublishOptions_HeaderOnEmptyOptions(t *testing.T) {
	opts := &broker.PublishOptions{}
	opts.Apply(broker.Header("k", "v"))

	if opts.Headers["k"] != "v" {
		t.Errorf("Headers[%q] = %q, want %q", "k", opts.Headers["k"], "v")
	}
}
