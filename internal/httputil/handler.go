package httputil

import (
	"net/http"

	"golang.org/x/xerrors"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func NewHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			panic(xerrors.Errorf("http handler error: %w", err))
		}
	}
}
