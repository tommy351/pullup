package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nameField      = "spec.repositories.githubName"
	repoTypeGitHub = "github"
)

// HandlerSet provides a handler.
// nolint: gochecknoglobals
var HandlerSet = wire.NewSet(
	wire.Struct(new(Handler), "*"),
)

type Config struct {
	Secret string `mapstructure:"secret"`
}

type Handler struct {
	Config                 Config
	Client                 client.Client
	Recorder               record.EventRecorder
	Indexer                client.FieldIndexer
	ResourceTemplateClient hookutil.ResourceTemplateClient
}

func (h *Handler) Initialize() error {
	if err := h.buildIndexAlpha(); err != nil {
		return err
	}

	if err := h.buildIndexBeta(); err != nil {
		return err
	}

	return nil
}

func (h *Handler) buildIndexAlpha() error {
	return h.Indexer.IndexField(context.TODO(), &v1alpha1.Webhook{}, nameField, func(obj runtime.Object) []string {
		var result []string

		for _, repo := range obj.(*v1alpha1.Webhook).Spec.Repositories {
			if repo.Type == repoTypeGitHub {
				result = append(result, repo.Name)
			}
		}

		return result
	})
}

func (h *Handler) buildIndexBeta() error {
	return h.Indexer.IndexField(context.TODO(), &v1beta1.GitHubWebhook{}, nameField, func(obj runtime.Object) []string {
		var result []string

		for _, repo := range obj.(*v1beta1.GitHubWebhook).Spec.Repositories {
			result = append(result, repo.Name)
		}

		return result
	})
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	logger := logr.FromContextOrDiscard(r.Context())
	payload, err := h.parsePayload(r)
	if err != nil {
		logger.Error(err, "Invalid payload")

		return httputil.JSON(w, http.StatusBadRequest, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Invalid payload"},
			},
		})
	}

	if err := h.handlePayload(r, payload); err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, &httputil.Response{})
}

func (h *Handler) handlePayload(r *http.Request, payload interface{}) error {
	switch event := payload.(type) {
	case *github.PushEvent:
		return h.handlePushEvent(r.Context(), event)
	case *github.PullRequestEvent:
		return h.handlePullRequestEvent(r.Context(), event)
	}

	return nil
}

func (h *Handler) parsePayload(r *http.Request) (interface{}, error) {
	payload, err := github.ValidatePayload(r, []byte(h.Config.Secret))
	if err != nil {
		return nil, fmt.Errorf("invalid github payload: %w", err)
	}

	return github.ParseWebHook(github.WebHookType(r), payload)
}
