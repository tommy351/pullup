package middleware

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
)

func Recovery(handler func(w http.ResponseWriter, r *http.Request, err error)) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if e, ok := err.(error); ok {
						handler(w, r, e)
					} else {
						handler(w, r, fmt.Errorf("recovered from panic: %+v", err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
