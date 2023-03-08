package kafka

import (
	"context"
	"fmt"

	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
	"github.com/segmentio/kafka-go"
)

type UserEventsPublisher struct {
	writer     *kafka.Writer
	serializer Serializer
}

func NewUserEventsPublisher(writer *kafka.Writer, serializer Serializer) *UserEventsPublisher {
	return &UserEventsPublisher{writer: writer, serializer: serializer}
}

func (u *UserEventsPublisher) Publish(event application.Event) error {
	em, err := u.serializer.Serialize(event)
	if err != nil {
		return fmt.Errorf("error while serializing event: %w", err)
	}

	err = u.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: em,
			Headers: []kafka.Header{
				{Key: "eventType", Value: []byte(event.GetEventType())},
				{Key: "requestUUID", Value: []byte(event.GetRequestUUID().String())},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error while publishing event: %w", err)
	}

	return nil
}

type Serializer interface {
	Serialize(any any) ([]byte, error)
}
