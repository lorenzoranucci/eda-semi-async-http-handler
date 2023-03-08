package async

import "github.com/google/uuid"

type Request interface {
	GetRequestUUID() uuid.UUID
}

type MessageDeserializer interface {
	DeserializeMessage(message []byte, eventType string) (any, error)
}
