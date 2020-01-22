package github

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v29/github"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	nameField      = "spec.repositories.githubName"
	repoTypeGitHub = "github"
)

const (
	ReasonCreated        = "Created"
	ReasonCreateFailed   = "CreateFailed"
	ReasonUpdated        = "Updated"
	ReasonUpdateFailed   = "UpdateFailed"
	ReasonDeleted        = "Deleted"
	ReasonDeleteFailed   = "DeleteFailed"
	ReasonInvalidWebhook = "InvalidWebhook"
)

type Config struct {
	Secret string `mapstructure:"secret"`
}

type Handler struct {
	secret   string
	client   client.Client
	handler  http.Handler
	recorder record.EventRecorder
}

func NewHandler(conf Config, mgr manager.Manager) (*Handler, error) {
	err := mgr.GetFieldIndexer().IndexField(&v1alpha1.Webhook{}, nameField, func(obj runtime.Object) []string {
		var result []string

		for _, repo := range obj.(*v1alpha1.Webhook).Spec.Repositories {
			if repo.Type == repoTypeGitHub {
				result = append(result, repo.Name)
			}
		}

		return result
	})

	if err != nil {
		return nil, fmt.Errorf("failed to index field: %w", err)
	}

	h := &Handler{
		secret:   conf.Secret,
		client:   mgr.GetClient(),
		recorder: mgr.GetEventRecorderFor("pullup-webhook"),
	}

	h.handler = httputil.NewHandler(h.handle)
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) error {
	payload, err := h.parsePayload(r)

	if err != nil {
		return httputil.String(w, http.StatusBadRequest, "Invalid request")
	}

	if event, ok := payload.(*github.PullRequestEvent); ok {
		list := new(v1alpha1.WebhookList)
		err = h.client.List(r.Context(), list, client.MatchingFields(map[string]string{
			nameField: event.Repo.GetFullName(),
		}))

		if err != nil {
			return fmt.Errorf("failed to find matching webhooks: %w", err)
		}

		for _, hook := range list.Items {
			hook := hook
			hook.SetGroupVersionKind(k8s.Kind("Webhook"))

			if err := h.handlePullRequestEvent(r.Context(), event, &hook); err != nil {
				return fmt.Errorf("failed to handle pull request event: %w", err)
			}
		}
	}

	return httputil.NoContent(w)
}

func (h *Handler) parsePayload(r *http.Request) (interface{}, error) {
	payload, err := github.ValidatePayload(r, []byte(h.secret))

	if err != nil {
		return nil, err
	}

	return github.ParseWebHook(github.WebHookType(r), payload)
}
