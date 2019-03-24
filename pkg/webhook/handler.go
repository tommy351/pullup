package webhook

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
	"golang.org/x/xerrors"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func NewHandler(handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			panic(xerrors.Errorf("http handler error: %w", err))
		}
	})
}

func Params(r *http.Request) map[string]string {
	return httptreemux.ContextParams(r.Context())
}
