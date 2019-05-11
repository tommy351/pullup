package middleware

import (
	"net/http"

	"github.com/justinas/alice"
	"golang.org/x/xerrors"
)

func Recovery(handler func(w http.ResponseWriter, r *http.Request, err error)) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if e, ok := err.(error); ok {
						handler(w, r, e)
					} else {
						handler(w, r, xerrors.Errorf("recovered from panic: %w", err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
