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
	"github.com/tommy351/pullup/internal/jsonschema"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"github.com/xeipuuv/gojsonschema"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=httpwebhooks,verbs=get;list;watch

// HandlerSet provides a handler.
// nolint: gochecknoglobals
var HandlerSet = wire.NewSet(
	wire.Struct(new(Handler), "*"),
)

// nolint: gochecknoglobals
var bodySchemaLoader = gojsonschema.NewGoLoader(map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"namespace": map[string]interface{}{
			"type":   "string",
			"format": "kubernetes_namespace",
		},
		"name": map[string]interface{}{
			"type":   "string",
			"format": "kubernetes_name",
		},
		"action": map[string]interface{}{
			"type": "string",
			"enum": []interface{}{hookutil.ActionApply, hookutil.ActionDelete},
		},
	},
	"requiredKeys": []string{"namespace", "name", "action"},
})

type Body struct {
	Namespace string                          `json:"namespace"`
	Name      string                          `json:"name"`
	Action    hookutil.ResourceTemplateAction `json:"action"`
	Data      extv1.JSON                      `json:"data"`
}

type Handler struct {
	Client                 client.Client
	ResourceTemplateClient hookutil.ResourceTemplateClient
}

func (h *Handler) parseBody(r *http.Request) (*Body, []httputil.Error, error) {
	logger := logr.FromContextOrDiscard(r.Context())

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return nil, []httputil.Error{
			{Description: `Content type must be "application/json"`},
		}, nil
	}

	var body Body

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error(err, "invalid json")

		return nil, []httputil.Error{
			{Description: "Invalid JSON"},
		}, nil
	}

	docLoader := gojsonschema.NewGoLoader(&body)
	result, err := jsonschema.Validate(bodySchemaLoader, docLoader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate body: %w", err)
	}

	if !result.Valid() {
		return nil, httputil.NewErrorsForJSONSchema(result.Errors()), nil
	}

	return &body, nil, nil
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	body, bodyErrors, err := h.parseBody(r)
	if err != nil {
		return err
	}

	if len(bodyErrors) > 0 {
		return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
			Errors: bodyErrors,
		})
	}

	hook := new(v1beta1.HTTPWebhook)
	err = h.Client.Get(r.Context(), types.NamespacedName{
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

	docLoader := gojsonschema.NewBytesLoader(body.Data.Raw)

	if schema := hook.Spec.Schema; schema != nil {
		schemaLoader := gojsonschema.NewBytesLoader(schema.Raw)
		result, err := jsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			return fmt.Errorf("failed to validate data: %w", err)
		}

		if !result.Valid() {
			return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
				Errors: httputil.NewErrorsForJSONSchema(result.Errors()),
			})
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
