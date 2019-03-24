package middleware

import (
	"context"
	"net/http"

	"github.com/justinas/alice"
)

func Context(ctx context.Context) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
