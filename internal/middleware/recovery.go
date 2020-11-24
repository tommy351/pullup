package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Recovery(handler func(w http.ResponseWriter, r *http.Request, err error)) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if e, ok := err.(error); ok {
						handler(w, r, e)
					} else {
						handler(w, r, fmt.Errorf("recovered from panic: %w", err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
