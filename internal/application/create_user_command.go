package application

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type CreateUserCommand struct {
	UserName    string
	RequestUUID uuid.UUID
}

type CreateUserCommandHandler struct {
	repository      UserRepository
	eventsPublisher UserEventsPublisher
}

func NewCreateUserCommandHandler(repository UserRepository, eventsPublisher UserEventsPublisher) *CreateUserCommandHandler {
	return &CreateUserCommandHandler{repository: repository, eventsPublisher: eventsPublisher}
}

// Handle is idempotent and, in case of error, must be retried from its caller until it succeeds.
// This ensures that the command is executed at least once completely and eventually the event is published at least once.
func (h *CreateUserCommandHandler) Handle(command CreateUserCommand) error {
	u, err := h.repository.FindUserByName(command.UserName)
	if err != nil {
		return fmt.Errorf("failed to find user by name: %w", err)
	}

	if u != nil && u.RequestUUID != command.RequestUUID {
		return h.eventsPublisher.Publish(&UserDuplicatedEvent{
			HTTPRequestContext: HTTPRequestContext{
				RequestUUID: command.RequestUUID,
			},
			UserName: command.UserName,
		})
	}

	if u == nil {
		u = &User{
			UserUUID:    uuid.New(),
			UserName:    command.UserName,
			RequestUUID: command.RequestUUID,
		}
		err := h.repository.AddUser(u)

		if err != nil {
			return fmt.Errorf("failed to add user: %w", err)
		}
	}

	return h.eventsPublisher.Publish(&UserCreatedEvent{
		HTTPRequestContext: HTTPRequestContext{
			RequestUUID: command.RequestUUID,
		},
		UserUUID: u.UserUUID,
		UserName: command.UserName,
	})
}

type CreateUserRequestSentEvent struct {
	HTTPRequestContext
	UserName string `json:"name"`
}

type UserCreatedEvent struct {
	HTTPRequestContext
	UserUUID uuid.UUID `json:"userUUID,omitempty"`
	UserName string    `json:"name,omitempty"`
}

func (u *UserCreatedEvent) GetEventType() string {
	return "UserCreated"
}

type UserDuplicatedEvent struct {
	HTTPRequestContext
	UserName string `json:"name,omitempty"`
}

func (u *UserDuplicatedEvent) GetEventType() string {
	return "UserDuplicated"
}

type User struct {
	RequestUUID uuid.UUID
	UserUUID    uuid.UUID
	UserName    string
}

type UserRepository interface {
	AddUser(user *User) error
	FindUserByName(name string) (*User, error)
}

type Event interface {
	GetEventType() string
	GetRequestUUID() uuid.UUID
}

type UserEventsPublisher interface {
	Publish(event Event) error
}

type UserEventDeserializer struct{}

func (c *UserEventDeserializer) DeserializeMessage(message []byte, eventType string) (any, error) {
	switch eventType {
	case "UserCreated":
		var userCreatedEvent UserCreatedEvent
		return userCreatedEvent, json.Unmarshal(message, &userCreatedEvent)
	case "UserDuplicated":
		var userDuplicatedEvent UserDuplicatedEvent
		return userDuplicatedEvent, json.Unmarshal(message, &userDuplicatedEvent)
	}

	return nil, fmt.Errorf("unknown event type: %s", eventType)
}
