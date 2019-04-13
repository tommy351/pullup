package webhook

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	Client client.Client
	Logger logr.Logger
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	hook := new(v1alpha1.Webhook)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, hook); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get webhook: %w", err)
	}

	logger := r.Logger.WithValues("webhook", hook)
	ctx = log.NewContext(ctx, logger)

	var list v1alpha1.ResourceSetList
	err := r.Client.List(ctx, &list, client.MatchingLabels(map[string]string{
		k8s.LabelWebhookName: hook.Name,
	}))

	if err != nil {
		return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to list resource sets: %w", err)
	}

	for _, set := range list.Items {
		set := set
		if err := r.patchResourceSet(ctx, hook, &set); err != nil {
			return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to patch resource set: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) patchResourceSet(ctx context.Context, webhook *v1alpha1.Webhook, set *v1alpha1.ResourceSet) error {
	logger := log.FromContext(ctx).WithValues("resourceSet", set)

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

	if err := r.Client.Patch(ctx, set, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
		return xerrors.Errorf("failed to patch the resource set: %w", err)
	}

	logger.V(log.Debug).Info("Patched resource set")
	return nil
}
