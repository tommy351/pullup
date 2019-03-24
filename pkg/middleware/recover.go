package middleware

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
	"golang.org/x/xerrors"
)

func Recover(handler http.Handler) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger := hlog.FromRequest(r)
					logger.Error().Stack().
						Err(xerrors.Errorf("http handler panicked: %w", err)).
						Msg("HTTP handler panicked")

					handler.ServeHTTP(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
