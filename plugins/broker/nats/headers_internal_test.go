package nats

import "testing"

func TestNatsHeaderFrom_ConvertsMap(t *testing.T) {
	got := natsHeaderFrom(map[string]string{"trace-id": "abc"})
	if got.Get("trace-id") != "abc" {
		t.Errorf("natsHeaderFrom()[trace-id] = %q, want %q", got.Get("trace-id"), "abc")
	}
}

func TestNatsHeaderFrom_Empty(t *testing.T) {
	if got := natsHeaderFrom(nil); got != nil {
		t.Errorf("natsHeaderFrom(nil) = %v, want nil", got)
	}
}
