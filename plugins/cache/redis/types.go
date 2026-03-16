package redis

import (
	"encoding/json"
	"fmt"
)

type (
	KeyFunc[K comparable] func(K) string

	Codec[V any] interface {
		Marshal(V) ([]byte, error)
		Unmarshal([]byte, *V) error
	}

	JSONCodec[V any] struct{}

	BytesCodec struct{}
)

func DefaultKeyFunc[K comparable](k K) string {
	return fmt.Sprint(k)
}

func (JSONCodec[V]) Marshal(v V) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec[V]) Unmarshal(data []byte, v *V) error {
	return json.Unmarshal(data, v)
}

func (BytesCodec) Marshal(v []byte) ([]byte, error) {
	return v, nil
}

func (BytesCodec) Unmarshal(data []byte, v *[]byte) error {
	*v = append((*v)[:0], data...)
	return nil
}
