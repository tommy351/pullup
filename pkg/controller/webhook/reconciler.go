package webhook

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ReasonPatched     = "Patched"
	ReasonPatchFailed = "PatchFailed"
)

type Reconciler struct {
	EventRecorder record.EventRecorder

	client client.Client
	logger logr.Logger
}

func (r *Reconciler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

func (r *Reconciler) InjectLogger(l logr.Logger) error {
	r.logger = l
	return nil
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	hook := new(v1alpha1.Webhook)
	ctx := context.Background()

	if err := r.client.Get(ctx, req.NamespacedName, hook); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get webhook: %w", err)
	}

	logger := r.logger.WithValues("webhook", hook)
	ctx = log.NewContext(ctx, logger)

	list := new(v1alpha1.ResourceSetList)
	err := r.client.List(ctx, list, client.MatchingLabels(map[string]string{
		k8s.LabelWebhookName: hook.Name,
	}))

	if err != nil {
		return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to list resource sets: %w", err)
	}

	for _, set := range list.Items {
		set := set

		if err := r.patchResourceSet(ctx, hook, &set); err != nil {
			r.EventRecorder.Eventf(hook, corev1.EventTypeWarning, ReasonPatchFailed, "Failed to patch resource set %q: %v", set.Name, err)
			return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to patch resource set: %w", err)
		}

		r.EventRecorder.Eventf(hook, corev1.EventTypeNormal, ReasonPatched, "Patched resource set %q", set.Name)
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

	if err := r.client.Patch(ctx, set, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
		return xerrors.Errorf("failed to patch the resource set: %w", err)
	}

	logger.V(log.Debug).Info("Patched resource set")
	return nil
}
