package hookutil

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/santhosh-tekuri/jsonschema/v2"
	"github.com/tommy351/pullup/internal/httputil"
)

func formatJSONSchemaValidationError(err *jsonschema.ValidationError) (output []httputil.Error) {
	causes := err.Causes

	if len(causes) == 0 {
		causes = append(causes, err)
	}

	for _, cause := range causes {
		var field string

		if cause.InstancePtr != "#" {
			field = strings.TrimPrefix(cause.InstancePtr, "#/")
		}

		output = append(output, httputil.Error{
			Description: cause.Message,
			Field:       field,
		})
	}

	return
}

func NewHandler(handler httputil.Handler) http.Handler {
	return httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		logger := logr.FromContextOrDiscard(r.Context())

		if err := handler(w, r); err != nil {
			var (
				ve   ValidationErrors
				tnfe TriggerNotFoundError
				jsse *jsonschema.SchemaError
				jsve *jsonschema.ValidationError
			)

			switch {
			case errors.Is(err, ErrInvalidAction):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Invalid action"},
					},
				}

			case errors.As(err, &ve):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors:     httputil.NewValidationErrors("", ve),
				}

			case errors.As(err, &tnfe):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Trigger not found"},
					},
				}

			case errors.As(err, &jsse):
				logger.Error(err, "Invalid JSON schema")

				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors: []httputil.Error{
						{Description: "Invalid JSON schema"},
					},
				}

			case errors.As(err, &jsve):
				return httputil.Response{
					StatusCode: http.StatusBadRequest,
					Errors:     formatJSONSchemaValidationError(jsve),
				}

			default:
				return err
			}
		}

		return nil
	})
}
