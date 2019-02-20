package server

import (
	"encoding/json"
	"net/http"

	"github.com/ansel1/merry"
)

func JSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return merry.Wrap(json.NewEncoder(w).Encode(data))
}

func String(w http.ResponseWriter, status int, data string) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(data))
	return merry.Wrap(err)
}

func Error(w http.ResponseWriter, err *APIError) error {
	status := err.StatusCode

	if status == 0 {
		status = http.StatusInternalServerError
	}

	return JSON(w, status, err)
}
