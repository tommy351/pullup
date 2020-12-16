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
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	result.RecordEvent(h.Recorder, hook)

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
		},
		Spec: v1alpha1.ResourceSetSpec{
			Resources: hook.Spec.Resources,
			Number:    event.GetNumber(),
			Base:      branchToCommitAlpha(event.PullRequest.Base),
			Head:      branchToCommitAlpha(event.PullRequest.Head),
		},
	}

	if err := controllerutil.SetControllerReference(hook, rs, h.Client.Scheme()); err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to set controller reference: %w", err),
			Reason: hookutil.ReasonFailed,
		}
	}

	var err error

	if rs.Name, err = getResourceName(event, hook, rs); err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to generate resource name: %w", err),
			Reason: hookutil.ReasonInvalidWebhook,
		}
	}

	if err := h.Client.Create(ctx, rs); err == nil {
		return controller.Result{
			Message: fmt.Sprintf("Created resource set: %s", rs.Name),
			Reason:  hookutil.ReasonCreated,
		}
	} else if !errors.IsAlreadyExists(err) {
		return controller.Result{
			Error:  fmt.Errorf("failed to create resource set: %w", err),
			Reason: hookutil.ReasonCreateFailed,
		}
	}

	patchValue, err := json.Marshal(rs.Spec)
	if err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to marshal patch value: %w", err),
			Reason: hookutil.ReasonUpdateFailed,
		}
	}

	patch, err := json.Marshal([]v1beta1.JSONPatch{
		{
			Operation: v1beta1.JSONPatchOpReplace,
			Path:      "/spec",
			Value:     &extv1.JSON{Raw: patchValue},
		},
	})
	if err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to marshal resource set spec: %w", err),
			Reason: hookutil.ReasonUpdateFailed,
		}
	}

	if err := h.Client.Patch(ctx, rs, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to patch resource set: %w", err),
			Reason: hookutil.ReasonUpdateFailed,
		}
	}

	return controller.Result{
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
			Error:  fmt.Errorf("failed to delete resource set: %w", err),
			Reason: hookutil.ReasonDeleteFailed,
		}
	}

	return controller.Result{
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
