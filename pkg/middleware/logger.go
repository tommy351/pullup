package middleware

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/justinas/alice"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/zenazn/goji/web/mutil"
)

func SetLogger(logger logr.Logger) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := log.NewContext(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestLog(log func(r *http.Request, status, size int, duration time.Duration)) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := mutil.WrapWriter(w)
			defer log(r, lw.Status(), lw.BytesWritten(), time.Since(start))
			next.ServeHTTP(lw, r)
		})
	}
}
