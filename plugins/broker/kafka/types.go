package kafka

import (
	"context"
	"log/slog"

	json "github.com/bytedance/sonic"
	"github.com/pthethanh/nano/broker"
)

type (
	logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}

	// Codec defines the interface gRPC uses to encode and decode messages.  Note
	// that implementations of this interface must be thread safe; a Codec's
	// methods can be called from concurrent goroutines.
	codec interface {
		// Marshal returns the wire format of v.
		Marshal(v any) ([]byte, error)
		// Unmarshal parses the wire format into v.
		Unmarshal(data []byte, v any) error
	}

	PublishError struct {
		Error   error
		Message *broker.Message
	}

	JSONCodec struct{}
)

func (m JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
