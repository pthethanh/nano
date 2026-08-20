package server

import (
	"testing"

	"google.golang.org/grpc/test/bufconn"
)

// TestUseSingleAddress_SameListenerObject proves that passing the same
// listener object for both grpc and http traffic (as the Listener() option
// does) is detected as single-address mode even when the listener's Addr()
// isn't a parseable host:port string (e.g. bufconn's Addr().String() ==
// "bufconn"). Without this, Listener(bufconnLis) would incorrectly run in
// separate-address mode with grpc and http racing to Accept() on the same
// underlying listener.
func TestUseSingleAddress_SameListenerObject(t *testing.T) {
	lis := bufconn.Listen(1024)
	srv := New(Listener(lis))
	if !srv.useSingleAddress() {
		t.Error("useSingleAddress() = false, want true for Listener(lis) (same object for grpc and http)")
	}
}

func TestUseSingleAddress_DifferentListenerObjects(t *testing.T) {
	a := bufconn.Listen(1024)
	b := bufconn.Listen(1024)
	srv := New(SeparateListeners(a, b))
	if srv.useSingleAddress() {
		t.Error("useSingleAddress() = true, want false for SeparateListeners(a, b) (different objects)")
	}
}
