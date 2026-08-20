package nats

import "testing"

func TestAddress_AcceptsVariadicAddresses(t *testing.T) {
	n := New[string](Address[string]("a:1", "b:2"))
	if len(n.addrs) != 2 || n.addrs[0] != "a:1" || n.addrs[1] != "b:2" {
		t.Errorf("addrs = %v, want [a:1 b:2]", n.addrs)
	}
}
