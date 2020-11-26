package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=webhooks,verbs=get;list;watch
// +kubebuilder:rbac:groups=pullup.dev,resources=resourcesets,verbs=patch

const (
	ReasonPatched     = "Patched"
	ReasonPatchFailed = "PatchFailed"
)

// ReconcilerSet provides a reconciler.
// nolint: gochecknoglobals
var ReconcilerSet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Reconciler), "*"),
)

type Logger logr.Logger

func NewLogger(logger logr.Logger) Logger {
	return logger.WithName("controller").WithName("webhook")
}

type Reconciler struct {
	Client   client.Client
	Logger   Logger
	Recorder record.EventRecorder
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	hook := new(v1alpha1.Webhook)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, hook); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get webhook: %w", err)
	}

	logger := r.Logger.WithValues("webhook", hook)
	ctx = logr.NewContext(ctx, logger)

	list := new(v1alpha1.ResourceSetList)
	err := r.Client.List(ctx, list, client.InNamespace(hook.Namespace), client.MatchingLabels(map[string]string{
		k8s.LabelWebhookName: hook.Name,
	}))
	if err != nil {
		return reconcile.Result{Requeue: true}, fmt.Errorf("failed to list resource sets: %w", err)
	}

	for _, set := range list.Items {
		set := set
		result := r.patchResourceSet(ctx, hook, &set)

		result.RecordEvent(r.Recorder)

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

	if err := r.Client.Patch(ctx, set, client.RawPatch(types.JSONPatchType, patch)); err != nil {
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
