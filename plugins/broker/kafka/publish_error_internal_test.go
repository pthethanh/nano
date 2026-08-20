package kafka

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
)

// White-box test (same package) for the async-publish-failure mapping logic,
// exercised directly against a plain sarama.ProducerError (no producer or
// broker needed) so it can prove the bug without any network I/O.
func TestPublishErrorFrom_PreservesTypedFailedMessage(t *testing.T) {
	type myMsg struct{ ID string }
	original := &myMsg{ID: "abc"}
	srcErr := errors.New("boom")

	perr := &sarama.ProducerError{
		Err: srcErr,
		Msg: &sarama.ProducerMessage{Metadata: original},
	}

	got := publishErrorFrom[myMsg](perr)

	if got.Error != srcErr {
		t.Errorf("Error = %v, want %v", got.Error, srcErr)
	}
	if got.Message != original {
		t.Errorf("Message = %v, want %v: the failed message must be recoverable from the failure callback for any T, not just T == broker.Message", got.Message, original)
	}
}
