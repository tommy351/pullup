package github

import (
	"context"
	"fmt"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type webhookList struct {
	V1Alpha1 v1alpha1.WebhookList
	V1Beta1  v1beta1.GitHubWebhookList
}

func (h *Handler) listWebhooks(ctx context.Context, name string) (*webhookList, error) {
	list := webhookList{}
	options := []client.ListOption{
		client.MatchingFields(map[string]string{
			nameField: name,
		}),
	}
	targets := []runtime.Object{
		&list.V1Alpha1,
		&list.V1Beta1,
	}

	for _, target := range targets {
		if err := h.Client.List(ctx, target, options...); err != nil {
			return nil, fmt.Errorf("failed to list webhooks: %w", err)
		}
	}

	for _, item := range list.V1Alpha1.Items {
		item.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("Webhook"))
	}

	for _, item := range list.V1Beta1.Items {
		item.SetGroupVersionKind(v1beta1.GroupVersion.WithKind("GitHubWebhook"))
	}

	return &list, nil
}
