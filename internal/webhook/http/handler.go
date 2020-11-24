package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"github.com/xeipuuv/gojsonschema"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HandlerSet provides a handler.
// nolint: gochecknoglobals
var HandlerSet = wire.NewSet(
	wire.Struct(new(Handler), "*"),
)

type Body struct {
	Namespace string                          `json:"namespace"`
	Name      string                          `json:"name"`
	Action    hookutil.ResourceTemplateAction `json:"action"`
	Data      interface{}                     `json:"data"`
}

type Handler struct {
	Client                 client.Client
	ResourceTemplateClient hookutil.ResourceTemplateClient
}

func (h *Handler) Initialize() error {
	return nil
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	logger := logr.FromContextOrDiscard(r.Context())

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
			Errors: []httputil.Error{
				{Description: `Content type must be "application/json"`},
			},
		})
	}

	var body Body

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error(err, "invalid json")

		return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Invalid JSON"},
			},
		})
	}

	if body.Name == "" {
		return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Resource name is required", Field: "name"},
			},
		})
	}

	hook := new(v1beta1.HTTPWebhook)
	err := h.Client.Get(r.Context(), types.NamespacedName{
		Namespace: body.Namespace,
		Name:      body.Name,
	}, hook)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
				Errors: []httputil.Error{
					{Description: "HTTPWebhook not found"},
				},
			})
		}

		return fmt.Errorf("failed to get HTTPWebhook: %w", err)
	}

	if schema := hook.Spec.Schema; schema != nil {
		schemaLoader := gojsonschema.NewBytesLoader(schema.Raw)
		docLoader := gojsonschema.NewGoLoader(body.Data)
		result, err := gojsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			return fmt.Errorf("failed to validate data: %w", err)
		}

		if !result.Valid() {
			var res httputil.Response

			for _, err := range result.Errors() {
				res.Errors = append(res.Errors, httputil.Error{
					Description: err.Description(),
					Field:       err.Field(),
				})
			}

			return httputil.JSON(w, http.StatusBadRequest, &res)
		}
	}

	err = h.ResourceTemplateClient.Handle(r.Context(), &hookutil.ResourceTemplateOptions{
		Action:  body.Action,
		Event:   body.Data,
		Webhook: hook,
	})
	if err != nil {
		if errors.Is(err, hookutil.ErrInvalidResourceTemplateAction) {
			return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
				Errors: []httputil.Error{
					{Description: "Invalid action", Field: "action"},
				},
			})
		}

		return fmt.Errorf("failed to %s resource template: %w", body.Action, err)
	}

	return httputil.JSON(w, http.StatusOK, &httputil.Response{})
}
