package event

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WebhookReconciler struct {
	Client client.Client
	Logger zerolog.Logger
}

func (w *WebhookReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	var hook v1alpha1.Webhook
	logger := w.Logger.With().
		Dict("webhook", zerolog.Dict().
			Str("namespace", req.Namespace).
			Str("name", req.Name)).
		Logger()
	ctx := logger.WithContext(context.Background())

	if err := w.Client.Get(ctx, req.NamespacedName, &hook); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get webhook: %w", err)
	}

	var list v1alpha1.ResourceSetList
	err := w.Client.List(ctx, &list, client.MatchingLabels(map[string]string{
		k8s.LabelWebhookName: hook.Name,
	}))

	if err != nil {
		return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to list resource sets: %w", err)
	}

	for _, set := range list.Items {
		set := set
		if err := w.patchResourceSet(ctx, &hook, &set); err != nil {
			return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to patch resource set: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (w *WebhookReconciler) patchResourceSet(ctx context.Context, webhook *v1alpha1.Webhook, set *v1alpha1.ResourceSet) error {
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

	if err := w.Client.Patch(ctx, set, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
		return xerrors.Errorf("failed to patch the resource set: %w", err)
	}

	logger.Debug().Msg("Patched resource set")
	return nil
}
