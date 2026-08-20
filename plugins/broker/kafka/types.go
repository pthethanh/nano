package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

type (
	logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}

	// PublishError carries the failure and, when recoverable, the original
	// message for an async publish failure. T matches the Broker's message
	// type, so Message is always populated when the wrapped sarama error
	// carries metadata (see OnAsyncPublishFailure).
	PublishError[T any] struct {
		Error   error
		Message *T
	}

	JSONCodec[T any] struct{}
)

func (m JSONCodec[T]) Marshal(v *T) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONCodec[T]) Unmarshal(data []byte, v *T) error {
	return json.Unmarshal(data, v)
}

// recordHeadersFrom converts broker.PublishOptions.Headers into sarama
// record headers.
func recordHeadersFrom(headers map[string]string) []sarama.RecordHeader {
	if len(headers) == 0 {
		return nil
	}
	out := make([]sarama.RecordHeader, 0, len(headers))
	for k, v := range headers {
		out = append(out, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
	}
	return out
}

// publishErrorFrom maps a sarama async-producer error to a PublishError[T],
// recovering the original message from the producer message's Metadata
// (set in Broker.Publish) when present.
func publishErrorFrom[T any](err *sarama.ProducerError) *PublishError[T] {
	pErr := &PublishError[T]{Error: err.Err}
	if err.Msg != nil && err.Msg.Metadata != nil {
		if v, ok := err.Msg.Metadata.(*T); ok {
			pErr.Message = v
		}
	}
	return pErr
}
