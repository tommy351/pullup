package middleware

import (
	"net/http"

	"github.com/ansel1/merry"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Recover(handler http.Handler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger := hlog.FromRequest(r)
					event := logger.Error()

					if err, ok := err.(error); ok {
						event = event.Stack().Err(merry.Wrap(err))
					} else {
						event = event.Interface(zerolog.ErrorFieldName, err)
					}

					event.Msg("")
					handler.ServeHTTP(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
