package github

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
)

func (h *Handler) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent) error {
	repoName := event.Repo.GetFullName()
	list, err := h.listWebhooks(ctx, repoName)
	if err != nil {
		return err
	}

	logger := logr.FromContextOrDiscard(ctx).WithValues(
		"repository", repoName,
		"action", event.GetAction(),
	)
	ctx = logr.NewContext(ctx, logger)

	for _, hook := range list.V1Beta1.Items {
		hook := hook

		if err := h.handlePullRequestEventBeta(ctx, event, &hook); err != nil {
			return fmt.Errorf("failed to handle pull request event: %w", err)
		}
	}

	for _, hook := range list.V1Alpha1.Items {
		hook := hook

		if err := h.handlePullRequestEventAlpha(ctx, event, &hook); err != nil {
			return fmt.Errorf("failed to handle pull request event: %w", err)
		}
	}

	return nil
}

func getPullRequestEventLabels(event *github.PullRequestEvent) (result []string) {
	if label := event.Label; label != nil && label.Name != nil {
		result = append(result, *label.Name)
	}

	if pr := event.PullRequest; pr != nil {
		result = append(result, getPullRequestLabels(pr)...)
	}

	return
}

func getPullRequestLabels(pr *github.PullRequest) (result []string) {
	for _, label := range pr.Labels {
		if label != nil && label.Name != nil {
			result = append(result, *label.Name)
		}
	}

	return
}
