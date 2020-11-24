package middleware

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/zenazn/goji/web/mutil"
)

func SetLogger(logger logr.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logr.NewContext(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestLog(log func(r *http.Request, status, size int, duration time.Duration)) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := mutil.WrapWriter(w)
			defer func() {
				log(r, lw.Status(), lw.BytesWritten(), time.Since(start))
			}()
			next.ServeHTTP(lw, r)
		})
	}
}
