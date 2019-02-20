package server

import "net/http"

type APIErrorCode string

const (
	APIErrorCodeUnknown            = "unknown"
	APIErrorCodeNotFound           = "notFound"
	APIErrorCodeRepositoryNotFound = "repositoryNotFound"
	APIErrorCodeInvalidPayload     = "invalidPayload"
)

var (
	ErrUnknown = &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       APIErrorCodeUnknown,
		Message:    "Unknown error",
	}

	ErrNotFound = &APIError{
		StatusCode: http.StatusNotFound,
		Code:       APIErrorCodeNotFound,
		Message:    "Not found",
	}

	ErrInvalidPayload = &APIError{
		StatusCode: http.StatusBadRequest,
		Code:       APIErrorCodeInvalidPayload,
		Message:    "Invalid payload",
	}

	ErrRepositoryNotFound = &APIError{
		StatusCode: http.StatusNotFound,
		Code:       APIErrorCodeRepositoryNotFound,
		Message:    "Repository not found",
	}
)

type APIError struct {
	Code    APIErrorCode `json:"error"`
	Message string       `json:"message"`

	StatusCode int `json:"-"`
}

func (a APIError) Error() string {
	return a.Message
}
