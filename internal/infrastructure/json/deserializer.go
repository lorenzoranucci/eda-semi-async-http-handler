package json

import "encoding/json"

type Deserializer[T any] struct{}

func (d *Deserializer[T]) Deserialize(m []byte) (T, error) {
	var v T
	return v, json.Unmarshal(m, &v)
}
