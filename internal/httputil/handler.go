package httputil

import (
	"errors"
	"fmt"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func NewHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var res Response

			if errors.As(err, &res) {
				status := res.StatusCode
				if status == 0 {
					status = http.StatusOK
				}

				_ = JSON(w, status, &res)
			} else {
				panic(fmt.Errorf("http handler error: %w", err))
			}
		}
	}
}
