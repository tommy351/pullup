package manager

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned"
	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

type WebhookEventHandler struct {
	Client versioned.Interface
}

func (w *WebhookEventHandler) OnUpdate(ctx context.Context, obj interface{}) error {
	webhook, ok := obj.(*v1alpha1.Webhook)

	if !ok {
		return nil
	}

	sets, err := w.Client.PullupV1alpha1().ResourceSets(webhook.Namespace).List(metav1.ListOptions{
		LabelSelector: k8s.LabelWebhookName + "=" + webhook.Name,
	})

	if err != nil {
		return xerrors.Errorf("failed to find resource sets: %w", err)
	}

	for _, set := range sets.Items {
		set := set

		if err := w.patchResourceSet(ctx, webhook, &set); err != nil {
			return xerrors.Errorf("failed to patch reosurce set: %w", err)
		}
	}

	return nil
}

func (w *WebhookEventHandler) patchResourceSet(ctx context.Context, webhook *v1alpha1.Webhook, set *v1alpha1.ResourceSet) error {
	logger := zerolog.Ctx(ctx).With().
		Dict("resourceSet", zerolog.Dict().
			Str("name", set.Name).
			Str("namespace", set.Namespace)).
		Logger()

	patch, err := json.Marshal([]k8s.JSONPatch{
		{
			Op:    "replace",
			Path:  "/spec/resources",
			Value: webhook.Spec.Resources,
		},
	})

	if err != nil {
		return xerrors.Errorf("failed to marshal json patch: %w", err)
	}

	_, err = w.Client.PullupV1alpha1().ResourceSets(set.Namespace).Patch(set.Name, types.JSONPatchType, patch)

	if err != nil {
		return xerrors.Errorf("failed to patch the resource set: %w", err)
	}

	logger.Debug().Msg("Patched resource set")
	return nil
}

func (*WebhookEventHandler) OnDelete(ctx context.Context, obj interface{}) error {
	return nil
}
