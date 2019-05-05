package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/go-github/v24/github"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/httputil"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const nameField = "spec.repositories.githubName"

type Config struct {
	Secret string `mapstructure:"secret"`
}

type Handler struct {
	secret  string
	client  client.Client
	handler http.Handler
}

func NewHandler(conf Config, mgr manager.Manager) (*Handler, error) {
	err := mgr.GetFieldIndexer().IndexField(&v1alpha1.Webhook{}, nameField, func(obj runtime.Object) []string {
		var result []string

		for _, repo := range obj.(*v1alpha1.Webhook).Spec.Repositories {
			if repo.Type == "github" {
				result = append(result, repo.Name)
			}
		}

		return result
	})

	if err != nil {
		return nil, err
	}

	h := &Handler{
		secret: conf.Secret,
		client: mgr.GetClient(),
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
		err = h.client.List(r.Context(), list, client.MatchingField(nameField, event.Repo.GetFullName()))

		if err != nil {
			return xerrors.Errorf("failed to find matching webhooks: %w", err)
		}

		for _, hook := range list.Items {
			hook := hook
			hook.SetGroupVersionKind(k8s.Kind("Webhook"))

			if err := h.handlePullRequestEvent(r.Context(), event, &hook); err != nil {
				return xerrors.Errorf("failed to handle pull request event: %w", err)
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

func (h *Handler) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) error {
	name, err := getResourceName(event, hook)

	if err != nil {
		return xerrors.Errorf("failed to get resource name: %w", err)
	}

	rs := &v1alpha1.ResourceSet{
		TypeMeta: k8s.GVKToTypeMeta(k8s.Kind("ResourceSet")),
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hook.Namespace,
			Name:      name,
		},
	}
	logger := log.FromContext(ctx).WithValues("resourceSet", rs)

	switch event.GetAction() {
	case "opened", "reopened", "synchronize":
		rs.Labels = map[string]string{
			k8s.LabelWebhookName:       hook.Name,
			k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
		}
		rs.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         hook.APIVersion,
				Kind:               hook.Kind,
				Name:               hook.Name,
				UID:                hook.UID,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			},
		}
		rs.Spec = v1alpha1.ResourceSetSpec{
			Resources: hook.Spec.Resources,
			Number:    event.GetNumber(),
			Base:      branchToCommit(event.PullRequest.Base),
			Head:      branchToCommit(event.PullRequest.Head),
		}

		if err := h.client.Create(ctx, rs); err == nil {
			logger.V(log.Debug).Info("Created resource set")
			return nil
		} else if !errors.IsAlreadyExists(err) {
			return xerrors.Errorf("failed to create resource set: %w", err)
		}

		patch, err := json.Marshal([]k8s.JSONPatch{
			{
				Op:    "replace",
				Path:  "/spec",
				Value: rs.Spec,
			},
		})

		if err != nil {
			return xerrors.Errorf("failed to marshal resource set spec: %w", err)
		}

		if err := h.client.Patch(ctx, rs, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
			return xerrors.Errorf("failed to patch resource set: %w", err)
		}

		logger.V(log.Debug).Info("Updated resource set")

	case "closed":
		if err := h.client.Delete(ctx, rs); err != nil && !errors.IsNotFound(err) {
			return xerrors.Errorf("failed to delete resource set %q: %w", name, err)
		}

		logger.V(log.Debug).Info("Deleted resource set")
	}

	return nil
}

func branchToCommit(branch *github.PullRequestBranch) v1alpha1.Commit {
	if branch == nil {
		return v1alpha1.Commit{}
	}

	return v1alpha1.Commit{
		Ref: branch.GetRef(),
		SHA: branch.GetSHA(),
	}
}
