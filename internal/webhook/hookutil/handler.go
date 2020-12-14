package hookutil

import (
	"errors"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/internal/httputil"
)

func NewHandler(handler httputil.Handler) http.Handler {
	return httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		logger := logr.FromContextOrDiscard(r.Context())

		if err := handler(w, r); err != nil {
			var (
				jsve JSONSchemaValidationErrors
				ve   ValidationErrors
				jse  JSONSchemaValidateError
				tnfe TriggerNotFoundError
			)

			switch {
			case errors.Is(err, ErrInvalidAction):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Invalid action"},
					},
				}

			case errors.As(err, &jsve):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors:     httputil.NewErrorsForJSONSchema(jsve),
				}

			case errors.As(err, &ve):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors:     httputil.NewValidationErrors("", ve),
				}

			case errors.As(err, &jse):
				logger.Error(err, "JSON schema validation error")

				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Failed to validate against JSON schema"},
					},
				}

			case errors.As(err, &tnfe):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Trigger not found"},
					},
				}

			default:
				return err
			}
		}

		return nil
	})
}
