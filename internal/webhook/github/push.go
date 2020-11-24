package github

import (
	"context"

	"github.com/google/go-github/v32/github"
)

func (h *Handler) handlePushEvent(ctx context.Context, event *github.PushEvent) error {
	repoName := event.Repo.GetFullName()
	list, err := h.listWebhooks(ctx, repoName)
	if err != nil {
		return err
	}

	for _, hook := range list.V1Beta1.Items {
		hook := hook

		if err := h.handlePushEventBeta(ctx, event, &hook); err != nil {
			return err
		}
	}

	return nil
}
