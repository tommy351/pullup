package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) error {
	repo := extractWebhookRepository(hook, event.Repo.GetFullName())
	logger := log.FromContext(ctx).WithValues(
		"repository", event.Repo.FullName,
		"webhook", hook,
		"action", event.GetAction(),
	)

	if repo == nil {
		logger.V(log.Debug).Info("Repository does not exist in the webhook")
		return nil
	}

	if branch := event.PullRequest.Base.GetRef(); !filterWebhook(&repo.Branches, branch) {
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

	result.RecordEvent(h.recorder)

	if err := result.Error; err != nil {
		logger.Error(err, result.GetMessage())
		return err
	}

	logger.Info(result.GetMessage())
	return nil
}

func (h *Handler) applyResourceSet(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) controller.Result {
	rs := &v1alpha1.ResourceSet{
		TypeMeta: k8s.GVKToTypeMeta(k8s.Kind("ResourceSet")),
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
			Base:      branchToCommit(event.PullRequest.Base),
			Head:      branchToCommit(event.PullRequest.Head),
		},
	}

	var err error

	if rs.Name, err = getResourceName(event, hook, rs); err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to generate resource name: %w", err),
			Reason: ReasonInvalidWebhook,
		}
	}

	if err := h.client.Create(ctx, rs); err == nil {
		return controller.Result{
			Object:  hook,
			Message: fmt.Sprintf("Created resource set: %s", rs.Name),
			Reason:  ReasonCreated,
		}
	} else if !errors.IsAlreadyExists(err) {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to create resource set: %w", err),
			Reason: ReasonCreateFailed,
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
			Reason: ReasonUpdateFailed,
		}
	}

	if err := h.client.Patch(ctx, rs, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to patch resource set: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	return controller.Result{
		Object:  hook,
		Message: fmt.Sprintf("Updated resource set: %s", rs.Name),
		Reason:  ReasonUpdated,
	}
}

func (h *Handler) deleteResourceSets(ctx context.Context, event *github.PullRequestEvent, hook *v1alpha1.Webhook) controller.Result {
	list := new(v1alpha1.ResourceSetList)
	err := h.client.List(ctx, list,
		client.InNamespace(hook.Namespace),
		client.MatchingLabels(map[string]string{
			k8s.LabelWebhookName:       hook.Name,
			k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
		}))

	if err != nil {
		return controller.Result{
			Object: hook,
			Error:  fmt.Errorf("failed to list resource sets: %w", err),
			Reason: ReasonDeleteFailed,
		}
	}

	if len(list.Items) == 0 {
		return controller.Result{
			Object:  hook,
			Message: "No matching resource sets to delete",
			Reason:  ReasonDeleted,
		}
	}

	deleted := make([]string, len(list.Items))

	for i, item := range list.Items {
		item := item

		if err := h.client.Delete(ctx, &item); err != nil && !errors.IsNotFound(err) {
			return controller.Result{
				Object: hook,
				Error:  fmt.Errorf("failed to delete resource set: %w", err),
				Reason: ReasonDeleteFailed,
			}
		}

		deleted[i] = item.Name
	}

	return controller.Result{
		Object:  hook,
		Message: fmt.Sprintf("Deleted resource sets: %s", strings.Join(deleted, ", ")),
		Reason:  ReasonDeleted,
	}
}

func extractWebhookRepository(hook *v1alpha1.Webhook, name string) *v1alpha1.WebhookRepository {
	for _, r := range hook.Spec.Repositories {
		r := r

		if r.Type == repoTypeGitHub && r.Name == name {
			return &r
		}
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
