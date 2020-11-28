package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

type Response struct {
	Errors []Error `json:"errors,omitempty"`
}

type Error struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
}

func String(w http.ResponseWriter, status int, data string) error {
	w.WriteHeader(status)

	if _, err := w.Write([]byte(data)); err != nil {
		return fmt.Errorf("failed to write http response: %w", err)
	}

	return nil
}

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func JSON(w http.ResponseWriter, status int, data interface{}) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to write json to http response: %w", err)
	}

	return nil
}

func NewErrorsForJSONSchema(errors []gojsonschema.ResultError) (result []Error) {
	for _, err := range errors {
		result = append(result, Error{
			Type:        err.Type(),
			Description: err.Description(),
			Field:       err.Field(),
		})
	}

	return
}

func NewValidationErrors(field string, errors []string) (result []Error) {
	for _, err := range errors {
		result = append(result, Error{
			Description: err,
			Field:       field,
		})
	}

	return
}
