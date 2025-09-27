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
