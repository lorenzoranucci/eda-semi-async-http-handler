package kafka

import (
	"context"
	"fmt"

	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
	"github.com/segmentio/kafka-go"
)

type CreateUserRequestSentEventConsumer struct {
	reader       *kafka.Reader
	h            *application.CreateUserCommandHandler
	deserializer Deserializer
}

func NewCreateUserRequestSentEventConsumer(reader *kafka.Reader, h *application.CreateUserCommandHandler, deserializer Deserializer) *CreateUserRequestSentEventConsumer {
	return &CreateUserRequestSentEventConsumer{reader: reader, h: h, deserializer: deserializer}
}

// ConsumeMessages is long-running and blocks until an error occurs.
// In case of error, when restarting, it will start from the last committed offset.
func (c *CreateUserRequestSentEventConsumer) ConsumeMessages() error {
	for {
		m, err := c.reader.FetchMessage(context.Background())
		if err != nil {
			return fmt.Errorf("error while reading messages from kafka: %w", err)
		}

		e, err := c.deserializer.Deserialize(m.Value)
		if err != nil {
			return fmt.Errorf("error while unmarshalling CreateUserRequestSentEvent: %w", err)
		}

		err = c.h.Handle(application.CreateUserCommand{
			UserName:    e.UserName,
			RequestUUID: e.GetRequestUUID(),
		})
		if err != nil {
			return fmt.Errorf("error while handling CreateUserCommand: %w", err)
		}

		err = c.reader.CommitMessages(context.Background(), m)
		if err != nil {
			return fmt.Errorf("error while committing message: %w", err)
		}
	}
}

type Deserializer interface {
	Deserialize(any []byte) (application.CreateUserRequestSentEvent, error)
}
