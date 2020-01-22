package httputil

import (
	"fmt"
	"net/http"
)

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
