package server

import (
	"net/http"

	"github.com/ansel1/merry"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func NewHandler(handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			switch err := err.(type) {
			case APIError:
				_ = Error(w, &err)
			case *APIError:
				_ = Error(w, err)
			default:
				panic(merry.Wrap(err))
			}
		}
	})
}
