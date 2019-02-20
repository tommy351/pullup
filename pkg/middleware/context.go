package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

func Context(ctx context.Context) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
