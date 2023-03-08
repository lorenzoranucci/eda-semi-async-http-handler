package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/pkg/async"
)

type CreateUserHandler struct {
	AsyncRequestHandler asyncRequestHandler
}

func NewCreateUserHandler(asyncRequestHandler asyncRequestHandler) *CreateUserHandler {
	return &CreateUserHandler{AsyncRequestHandler: asyncRequestHandler}
}

type asyncRequestHandler interface {
	HandleAsyncRequest(r async.Request, responseChan chan any) error
}

func (h *CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestUUID := uuid.New()
	dr, err := h.deserializeRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseChan := make(chan any, 1)
	err = h.AsyncRequestHandler.HandleAsyncRequest(
		&application.CreateUserRequestSentEvent{
			HTTPRequestContext: application.HTTPRequestContext{
				RequestUUID: requestUUID,
			},
			UserName: dr.UserName,
		},
		responseChan,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case <-time.After(5 * time.Second):
		http.Error(w, "timeout", http.StatusGatewayTimeout)
		return
	case resp := <-responseChan:
		switch resp.(type) {
		case application.UserCreatedEvent:
			w.WriteHeader(http.StatusCreated)
			return
		case application.UserDuplicatedEvent:
			w.WriteHeader(http.StatusConflict)
			return
		case error:
			log.Print(resp.(error).Error())
			http.Error(w, "cannot process request", http.StatusInternalServerError)
			return
		default:
			http.Error(w, "cannot process request", http.StatusInternalServerError)
			return
		}
	}
}

type CreateUserRequest struct {
	UserName string `json:"name"`
}

func (h *CreateUserHandler) deserializeRequest(r *http.Request) (CreateUserRequest, error) {
	v := CreateUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return CreateUserRequest{}, fmt.Errorf("error while deserializing request: %w", err)
	}
	return v, nil
}
