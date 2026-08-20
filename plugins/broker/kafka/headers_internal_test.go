package kafka

import "testing"

func TestRecordHeadersFrom_ConvertsMap(t *testing.T) {
	got := recordHeadersFrom(map[string]string{"a": "1"})
	if len(got) != 1 || string(got[0].Key) != "a" || string(got[0].Value) != "1" {
		t.Errorf("recordHeadersFrom() = %+v, want a single {a:1} header", got)
	}
}

func TestRecordHeadersFrom_Empty(t *testing.T) {
	if got := recordHeadersFrom(nil); got != nil {
		t.Errorf("recordHeadersFrom(nil) = %+v, want nil", got)
	}
}
