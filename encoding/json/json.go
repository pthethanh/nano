package json

import (
	"github.com/bytedance/sonic"
	"github.com/pthethanh/nano/encoding"
)

type (
	codec struct{}
)

func init() {
	encoding.RegisterCodec(&codec{})
}

func (m *codec) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func (m *codec) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

func (m *codec) Name() string {
	return "json"
}
