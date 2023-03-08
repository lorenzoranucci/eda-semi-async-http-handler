package kafka

import (
	"context"
	"fmt"

	"github.com/lorenzoranucci/eda-semi-async-http-handler/pkg/async"
	"github.com/segmentio/kafka-go"
)

type AsyncRequestHandler struct {
	writer          *kafka.Writer
	serializer      Serializer
	addObservers    chan Observer
	removeObservers chan Observer
	execCtx         context.Context
}

func NewAsyncRequestHandler(
	writer *kafka.Writer,
	serializer Serializer,
	addObservers chan Observer,
	removeObservers chan Observer,
	execCtx context.Context,
) *AsyncRequestHandler {
	return &AsyncRequestHandler{
		writer:          writer,
		serializer:      serializer,
		addObservers:    addObservers,
		removeObservers: removeObservers,
		execCtx:         execCtx,
	}
}

type Observer struct {
	request      async.Request
	responseChan chan any
}

func (a *AsyncRequestHandler) HandleAsyncRequest(request async.Request, responseChan chan any) error {
	message, err := a.serializer.Serialize(request)
	if err != nil {
		return fmt.Errorf("error while getting message from request: %w", err)
	}

	j := Observer{
		request:      request,
		responseChan: responseChan,
	}
	select {
	case a.addObservers <- j:
	case <-a.execCtx.Done():
		return fmt.Errorf("error while adding observer: %w", a.execCtx.Err())
	}

	err = a.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: message,
			Headers: []kafka.Header{
				{Key: "requestUUID", Value: []byte(request.GetRequestUUID().String())},
			},
		},
	)
	if err != nil {
		select {
		case a.removeObservers <- j:
		case <-a.execCtx.Done():
			return fmt.Errorf("error while removing observer: %w + %w", a.execCtx.Err(), err)
		}
		return fmt.Errorf("error while writing message to kafka: %w", err)
	}

	return nil
}

type Serializer interface {
	Serialize(any any) ([]byte, error)
}
