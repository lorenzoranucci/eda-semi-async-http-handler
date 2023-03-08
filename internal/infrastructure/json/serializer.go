package json

import "encoding/json"

type Serializer struct {
}

func (s *Serializer) Serialize(any any) ([]byte, error) {
	return json.Marshal(any)
}
