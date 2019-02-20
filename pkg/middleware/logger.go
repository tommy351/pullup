package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
)

func Logger() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		c := alice.New(
			hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
				hlog.FromRequest(r).Debug().
					Int("status", status).
					Int("size", size).
					Dur("duration", duration).
					Msg("")
			}),
			hlog.MethodHandler("method"),
			hlog.URLHandler("url"),
			hlog.RemoteAddrHandler("remoteIp"),
			hlog.UserAgentHandler("userAgent"),
		)

		return c.Then(next)
	}
}
