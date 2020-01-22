package httputil

import (
	"fmt"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func NewHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			panic(fmt.Errorf("http handler error: %w", err))
		}
	}
}
