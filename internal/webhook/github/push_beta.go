package github

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/tommy351/pullup/internal/gitutil"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

func (h *Handler) handlePushEventBeta(ctx context.Context, event *github.PushEvent, hook *v1beta1.GitHubWebhook) error {
	repoName := event.Repo.GetFullName()
	repo := extractRepositoryBeta(hook, repoName)
	logger := logr.FromContextOrDiscard(ctx).WithValues(
		"repository", repoName,
		"webhook", hook,
	)

	if repo == nil {
		logger.V(log.Debug).Info("Repository does not exist in the webhook")

		return nil
	}

	if repo.Push == nil {
		logger.V(log.Debug).Info("Push event filter is not set")

		return nil
	}

	ref, ok := gitutil.ParseRef(event.GetRef())
	if !ok {
		logger.V(log.Debug).Info("Invalid ref", "ref", event.GetRef())

		return nil
	}

	switch ref.Type {
	case gitutil.RefTypeBranch:
		if (repo.Push.Branches == nil && repo.Push.Tags != nil) || !hookutil.FilterWebhook(repo.Push.Branches, []string{ref.Name}) {
			logger.V(log.Debug).Info("Skipped on this branch", "branch", ref.Name)

			return nil
		}

	case gitutil.RefTypeTag:
		if repo.Push.Tags == nil || !hookutil.FilterWebhook(repo.Push.Tags, []string{ref.Name}) {
			logger.V(log.Debug).Info("Skipped on this tag", "tag", ref.Name)

			return nil
		}

	default:
		logger.V(log.Debug).Info("Unsupported ref type", "refType", ref.Type)

		return nil
	}

	options := &hookutil.TriggerOptions{
		DefaultAction: v1beta1.ActionApply,
		Action:        hook.Spec.Action,
		Event:         event,
		Source:        hook,
		Triggers:      hook.Spec.Triggers,
	}

	return h.TriggerHandler.Handle(ctx, options)
}
