package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"github.com/xeipuuv/gojsonschema"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=httpwebhooks,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// HandlerSet provides a handler.
// nolint: gochecknoglobals
var HandlerSet = wire.NewSet(
	wire.Struct(new(Handler), "*"),
)

type Body struct {
	Namespace string         `json:"namespace"`
	Name      string         `json:"name"`
	Action    v1beta1.Action `json:"action"`
	Data      extv1.JSON     `json:"data"`
}

type Handler struct {
	Client         client.Client
	TriggerHandler hookutil.TriggerHandler
}

func (h *Handler) parseBody(r *http.Request) (*Body, error) {
	logger := logr.FromContextOrDiscard(r.Context())

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return nil, httputil.Response{
			StatusCode: http.StatusBadRequest,
			Errors: []httputil.Error{
				{Description: `Content type must be "application/json"`},
			},
		}
	}

	var body Body

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error(err, "invalid json")

		return nil, httputil.Response{
			StatusCode: http.StatusBadRequest,
			Errors: []httputil.Error{
				{Description: "Invalid JSON"},
			},
		}
	}

	if e := validation.ValidateNamespaceName(body.Namespace, false); len(e) > 0 {
		return nil, httputil.Response{
			StatusCode: http.StatusBadRequest,
			Errors:     httputil.NewValidationErrors("namespace", e),
		}
	}

	if e := validation.NameIsDNSSubdomain(body.Name, false); len(e) > 0 {
		return nil, httputil.Response{
			StatusCode: http.StatusBadRequest,
			Errors:     httputil.NewValidationErrors("name", e),
		}
	}

	if !v1beta1.IsActionValid(body.Action) {
		return nil, httputil.Response{
			StatusCode: http.StatusBadRequest,
			Errors: []httputil.Error{
				{Description: "Invalid action", Field: "action"},
			},
		}
	}

	return &body, nil
}

func (h *Handler) validateSecretToken(r *http.Request, hook *v1beta1.HTTPWebhook) error {
	if hook.Spec.SecretToken == nil || hook.Spec.SecretToken.SecretKeyRef == nil {
		return nil
	}

	header := r.Header.Get("Pullup-Webhook-Secret")
	ref := hook.Spec.SecretToken.SecretKeyRef
	secret := new(corev1.Secret)
	secretName := types.NamespacedName{
		Namespace: hook.Namespace,
		Name:      ref.Name,
	}

	if err := h.Client.Get(r.Context(), secretName, secret); err != nil {
		if kerrors.IsNotFound(err) {
			return httputil.Response{
				StatusCode: http.StatusForbidden,
				Errors: []httputil.Error{
					{Description: "Secret not found"},
				},
			}
		}

		return fmt.Errorf("failed to get secret: %w", err)
	}

	value, ok := secret.Data[ref.Key]
	if !ok {
		return httputil.Response{
			StatusCode: http.StatusForbidden,
			Errors: []httputil.Error{
				{Description: "Key does not contain in the secret"},
			},
		}
	}

	if header != string(value) {
		return httputil.Response{
			StatusCode: http.StatusForbidden,
			Errors: []httputil.Error{
				{Description: "Secret mismatch"},
			},
		}
	}

	return nil
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	logger := logr.FromContextOrDiscard(r.Context())
	body, err := h.parseBody(r)
	if err != nil {
		return err
	}

	hook := new(v1beta1.HTTPWebhook)
	err = h.Client.Get(r.Context(), types.NamespacedName{
		Namespace: body.Namespace,
		Name:      body.Name,
	}, hook)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return httputil.Response{
				StatusCode: http.StatusBadRequest,
				Errors: []httputil.Error{
					{Description: "HTTPWebhook not found"},
				},
			}
		}

		return fmt.Errorf("failed to get HTTPWebhook: %w", err)
	}

	if err := h.validateSecretToken(r, hook); err != nil {
		return err
	}

	if schema := hook.Spec.Schema; schema != nil && schema.Raw != nil {
		rawData := body.Data.Raw
		if rawData == nil {
			rawData = []byte("null")
		}

		docLoader := gojsonschema.NewBytesLoader(rawData)
		schemaLoader := gojsonschema.NewBytesLoader(schema.Raw)
		result, err := gojsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			logger.Error(err, "JSON schema validate error")

			return httputil.Response{
				StatusCode: http.StatusBadRequest,
				Errors: []httputil.Error{
					{Description: "Failed to validate against JSON schema"},
				},
			}
		}

		if !result.Valid() {
			return httputil.Response{
				StatusCode: http.StatusBadRequest,
				Errors:     httputil.NewErrorsForJSONSchema(result.Errors()),
			}
		}
	}

	err = h.TriggerHandler.Handle(r.Context(), &hookutil.TriggerOptions{
		Source:   hook,
		Triggers: hook.Spec.Triggers,
		Action:   body.Action,
		Event:    body.Data,
	})
	if err != nil {
		return fmt.Errorf("trigger failed: %w", err)
	}

	return httputil.JSON(w, http.StatusOK, &httputil.Response{})
}
