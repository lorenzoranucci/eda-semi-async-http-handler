package application

import "github.com/google/uuid"

type HTTPRequestContext struct {
	RequestUUID uuid.UUID `json:"requestUUID"`
}

func (e *HTTPRequestContext) GetRequestUUID() uuid.UUID {
	return e.RequestUUID
}
