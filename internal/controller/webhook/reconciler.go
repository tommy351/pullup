package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ReasonPatched     = "Patched"
	ReasonPatchFailed = "PatchFailed"
)

type Reconciler struct {
	client   client.Client
	logger   logr.Logger
	recorder record.EventRecorder
}

func NewReconciler(mgr manager.Manager, logger logr.Logger) *Reconciler {
	return &Reconciler{
		client:   mgr.GetClient(),
		logger:   logger.WithName("controller").WithName("webhook"),
		recorder: mgr.GetEventRecorderFor("pullup-controller"),
	}
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	hook := new(v1alpha1.Webhook)
	ctx := context.Background()

	if err := r.client.Get(ctx, req.NamespacedName, hook); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get webhook: %w", err)
	}

	logger := r.logger.WithValues("webhook", hook)
	ctx = log.NewContext(ctx, logger)

	list := new(v1alpha1.ResourceSetList)
	err := r.client.List(ctx, list, client.InNamespace(hook.Namespace), client.MatchingLabels(map[string]string{
		k8s.LabelWebhookName: hook.Name,
	}))

	if err != nil {
		return reconcile.Result{Requeue: true}, fmt.Errorf("failed to list resource sets: %w", err)
	}

	for _, set := range list.Items {
		set := set
		result := r.patchResourceSet(ctx, hook, &set)

		result.RecordEvent(r.recorder)

		if err := result.Error; err != nil {
			logger.Error(err, result.GetMessage())
			return reconcile.Result{Requeue: result.Requeue}, err
		}

		logger.Info(result.GetMessage())
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) patchResourceSet(ctx context.Context, webhook *v1alpha1.Webhook, set *v1alpha1.ResourceSet) controller.Result {
	patch, err := json.Marshal([]k8s.JSONPatch{
		{
			Op:    "replace",
			Path:  "/spec/resources",
			Value: webhook.Spec.Resources,
		},
	})

	if err != nil {
		return controller.Result{
			Object: webhook,
			Error:  fmt.Errorf("failed to marshal json patch: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	if err := r.client.Patch(ctx, set, client.ConstantPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object:  webhook,
			Error:   fmt.Errorf("failed to patch the resource set: %w", err),
			Reason:  ReasonPatchFailed,
			Requeue: true,
		}
	}

	return controller.Result{
		Object:  webhook,
		Message: fmt.Sprintf("Patched resource set %q", set.Name),
		Reason:  ReasonPatched,
	}
}
