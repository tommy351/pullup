package middleware

import (
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Logger(logger *zerolog.Logger) alice.Constructor {
	return func(next http.Handler) http.Handler {
		c := alice.New(
			hlog.NewHandler(*logger),
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
