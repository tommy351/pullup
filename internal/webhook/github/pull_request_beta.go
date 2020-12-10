package github

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

func (h *Handler) handlePullRequestEventBeta(ctx context.Context, event *github.PullRequestEvent, hook *v1beta1.GitHubWebhook) error {
	repoName := event.Repo.GetFullName()
	eventAction := event.GetAction()
	repo := extractRepositoryBeta(hook, repoName)
	logger := logr.FromContextOrDiscard(ctx).WithValues(
		"webhook", hook,
	)
	ctx = logr.NewContext(ctx, logger)

	if repo == nil {
		logger.V(log.Debug).Info("Repository does not exist in the webhook")

		return nil
	}

	if repo.PullRequest == nil {
		logger.V(log.Debug).Info("Pull request event filter is not set")

		return nil
	}

	if !filterByPullRequestType(repo.PullRequest.Types, eventAction) {
		logger.V(log.Debug).Info("Skipped for the action")

		return nil
	}

	if branch := event.PullRequest.Base.GetRef(); !hookutil.FilterWebhook(repo.PullRequest.Branches, branch) {
		logger.V(log.Debug).Info("Skipped on this branch", "branch", branch)

		return nil
	}

	options := &hookutil.ResourceTemplateOptions{
		Action:              v1beta1.WebhookActionApply,
		Event:               event,
		Webhook:             hook,
		DefaultResourceName: "{{ .webhook.metadata.name }}-{{ .event.number }}",
	}

	if eventAction == "closed" {
		options.Action = v1beta1.WebhookActionDelete
	}

	return h.ResourceTemplateClient.Handle(ctx, options)
}
