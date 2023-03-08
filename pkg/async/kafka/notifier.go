package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/pkg/async"
	"github.com/segmentio/kafka-go"
)

type ObserversNotifier struct {
	reader              *kafka.Reader
	messageDeserializer async.MessageDeserializer
	execCtx             context.Context
}

func NewObserversNotifier(
	reader *kafka.Reader,
	messageDeserializer async.MessageDeserializer,
	execCtx context.Context,
) *ObserversNotifier {
	return &ObserversNotifier{
		reader:              reader,
		messageDeserializer: messageDeserializer,
		execCtx:             execCtx,
	}
}

func (o *ObserversNotifier) ConsumeMessages(messages chan kafka.Message, errCh chan error) {
	for {
		m, err := o.reader.ReadMessage(o.execCtx)
		if err != nil {
			errCh <- fmt.Errorf("error while reading messages from kafka: %w", err)
			return
		}

		messages <- m
	}
}

func (o *ObserversNotifier) NotifyObservers(
	addObservers chan Observer,
	removeObservers chan Observer,
	messages chan kafka.Message,
) {
	observersChannels := make(map[uuid.UUID][]chan any)

	for {
		select {
		case j := <-addObservers:
			addObserver(observersChannels, j)
		case j := <-removeObservers:
			delete(observersChannels, j.request.GetRequestUUID())
		case m := <-messages:
			o.notifyObserversAboutMessage(m, observersChannels)
		case <-o.execCtx.Done():
			return
		}
	}
}

func (o *ObserversNotifier) notifyObserversAboutMessage(
	m kafka.Message,
	observersChannels map[uuid.UUID][]chan any,
) {
	var requestUUID uuid.UUID
	var eventType string
	for i, header := range m.Headers {
		if header.Key == "requestUUID" {
			var err error
			requestUUID, err = uuid.Parse(string(m.Headers[i].Value))
			if err != nil {
				log.Printf("error while parsing request uuid: %v", err)
				return
			}
		}

		if header.Key == "eventType" {
			eventType = string(m.Headers[i].Value)
		}
	}

	if requestUUID == uuid.Nil {
		log.Printf("cannot find request uuid")
		return
	}

	if eventType == "" {
		log.Printf("cannot find event type")
		return
	}

	dm, err := o.messageDeserializer.DeserializeMessage(m.Value, eventType)
	var chOut = dm
	if err != nil {
		chOut = err
	}

	if ch, ok := observersChannels[requestUUID]; ok {
		for _, c := range ch {
			c <- chOut
		}
		delete(observersChannels, requestUUID)
	}
}

func addObserver(observersChannels map[uuid.UUID][]chan any, j Observer) {
	if _, ok := observersChannels[j.request.GetRequestUUID()]; !ok {
		observersChannels[j.request.GetRequestUUID()] = []chan any{j.responseChan}
		return
	}
	observersChannels[j.request.GetRequestUUID()] = append(observersChannels[j.request.GetRequestUUID()], j.responseChan)
}
