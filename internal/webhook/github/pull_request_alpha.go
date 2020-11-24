package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) handlePullRequestEventAlpha(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) error {
	repo := extractRepositoryAlpha(hook, event.Repo.GetFullName())
	logger := logr.FromContextOrDiscard(ctx).WithValues(
		"webhook", hook,
	)

	if repo == nil {
		logger.V(log.Debug).Info("Repository does not exist in the webhook")

		return nil
	}

	if branch := event.PullRequest.Base.GetRef(); !filterWebhookAlpha(&repo.Branches, branch) {
		logger.V(log.Debug).Info("Skipped on this branch", "branch", branch)

		return nil
	}

	var result controller.Result

	switch event.GetAction() {
	case "opened", "reopened", "synchronize":
		result = h.applyResourceSet(ctx, event, hook)
	case "closed":
		result = h.deleteResourceSets(ctx, event, hook)
	default:
		return nil
	}

	result.RecordEvent(h.Recorder)

	if err := result.Error; err != nil {
		logger.Error(err, result.GetMessage())

		return err
	}

	logger.Info(result.GetMessage())

	return nil
}

func (h *Handler) applyResourceSet(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) controller.Result {
	rs := &v1alpha1.ResourceSet{
		TypeMeta: k8s.GVKToTypeMeta(v1alpha1.GroupVersion.WithKind("ResourceSet")),
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hook.Namespace,
			Labels: map[string]string{
				k8s.LabelWebhookName:       hook.Name,
				k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         hook.APIVersion,
					Kind:               hook.Kind,
					Name:               hook.Name,
					UID:                hook.UID,
					Controller:         pointer.BoolPtr(true),
					BlockOwnerDeletion: pointer.BoolPtr(true),
				},
			},
		},
		Spec: v1alpha1.ResourceSetSpec{
			Resources: hook.Spec.Resources,
			Number:    event.GetNumber(),
			Base:      branchToCommitAlpha(event.PullRequest.Base),
			Head:      branchToCommitAlpha(event.PullRequest.Head),
		},
	}

	var err error

	if rs.Name, err = getResourceName(event, hook, rs); err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to generate resource name: %w", err),
			Reason: hookutil.ReasonInvalidWebhook,
		}
	}

	if err := h.Client.Create(ctx, rs); err == nil {
		return controller.Result{
			Object:  hook,
			Message: fmt.Sprintf("Created resource set: %s", rs.Name),
			Reason:  hookutil.ReasonCreated,
		}
	} else if !errors.IsAlreadyExists(err) {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to create resource set: %w", err),
			Reason: hookutil.ReasonCreateFailed,
		}
	}

	patch, err := json.Marshal([]k8s.JSONPatch{
		{
			Op:    "replace",
			Path:  "/spec",
			Value: rs.Spec,
		},
	})
	if err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to marshal resource set spec: %w", err),
			Reason: hookutil.ReasonUpdateFailed,
		}
	}

	if err := h.Client.Patch(ctx, rs, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to patch resource set: %w", err),
			Reason: hookutil.ReasonUpdateFailed,
		}
	}

	return controller.Result{
		Object:  hook,
		Message: fmt.Sprintf("Updated resource set: %s", rs.Name),
		Reason:  hookutil.ReasonUpdated,
	}
}

func (h *Handler) deleteResourceSets(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) controller.Result {
	err := h.Client.DeleteAllOf(ctx, &v1alpha1.ResourceSet{},
		client.InNamespace(hook.Namespace),
		client.MatchingLabels(map[string]string{
			k8s.LabelWebhookName:       hook.Name,
			k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
		}))
	if err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to delete resource set: %w", err),
			Reason: hookutil.ReasonDeleteFailed,
		}
	}

	return controller.Result{
		Object:  hook,
		Message: "Deleted resource sets",
		Reason:  hookutil.ReasonDeleted,
	}
}

func extractRepositoryAlpha(hook *v1alpha1.Webhook, name string) *v1alpha1.WebhookRepository {
	for _, r := range hook.Spec.Repositories {
		r := r

		if r.Type == repoTypeGitHub && r.Name == name {
			return &r
		}
	}

	return nil
}

func branchToCommitAlpha(branch *github.PullRequestBranch) v1alpha1.Commit {
	if branch == nil {
		return v1alpha1.Commit{}
	}

	return v1alpha1.Commit{
		Ref: branch.GetRef(),
		SHA: branch.GetSHA(),
	}
}

func filterWebhookAlpha(filter *v1alpha1.WebhookFilter, text string) bool {
	if len(filter.Include) > 0 && !hookutil.FilterByConditions(filter.Include, text) {
		return false
	}

	if len(filter.Exclude) > 0 && hookutil.FilterByConditions(filter.Exclude, text) {
		return false
	}

	return true
}
