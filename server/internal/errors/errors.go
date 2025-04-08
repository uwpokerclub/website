package errors

import (
	"fmt"
	"net/http"
)

// APIErrorResponse is a custom error used for responses in the case of an API error.
type APIErrorResponse struct {
	// Code is the HTTP status code for the error.
	Code int `json:"code"`
	// Type is the name of the HTTP status code.
	Type string `json:"type"`
	// Message is a more descriptive error message.
	Message string `json:"message"`
}

// Error returns a string corresponding to the type of error and a detailed error message.
func (e APIErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func InvalidRequest(message string) error {
	return APIErrorResponse{
		Code:    http.StatusBadRequest,
		Type:    "INVALID_REQUEST",
		Message: message,
	}
}

func Unauthorized(message string) error {
	return APIErrorResponse{
		Code:    http.StatusUnauthorized,
		Type:    "UNAUTHORIZED",
		Message: message,
	}
}

func Forbidden(message string) error {
	return APIErrorResponse{
		Code:    http.StatusForbidden,
		Type:    "FORBIDDEN",
		Message: message,
	}
}

func NotFound(message string) error {
	return APIErrorResponse{
		Code:    http.StatusNotFound,
		Type:    "NOT_FOUND",
		Message: message,
	}
}

func InternalServerError(message string) error {
	return APIErrorResponse{
		Code:    http.StatusInternalServerError,
		Type:    "INTERNAL_SERVER_ERROR",
		Message: message,
	}
}
