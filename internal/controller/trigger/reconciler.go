package trigger

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=pullup.dev,resources=triggers,verbs=get;list;watch

const (
	ReasonPatched     = "Patched"
	ReasonPatchFailed = "PatchFailed"
)

const triggerRefField = "spec.triggerRef"

// ReconcilerConfigSet provides a ReconcilerConfig.
// nolint: gochecknoglobals
var ReconcilerConfigSet = wire.NewSet(
	NewLogger,
	wire.Struct(new(ReconcilerConfig), "*"),
)

// ReconcilerSet provides a Reconciler.
// nolint: gochecknoglobals
var ReconcilerSet = wire.NewSet(
	ReconcilerConfigSet,
	NewReconciler,
)

type Logger logr.Logger

func NewLogger(logger logr.Logger) Logger {
	return logger.WithName("controller").WithName("trigger")
}

type ReconcilerConfig struct {
	Client   client.Client
	Logger   Logger
	Recorder record.EventRecorder
}

type Reconciler struct {
	ReconcilerConfig
}

func NewReconciler(conf ReconcilerConfig, mgr manager.Manager) (*Reconciler, error) {
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.ResourceTemplate{}, triggerRefField, func(obj runtime.Object) []string {
		var result []string

		if ref := obj.(*v1beta1.ResourceTemplate).Spec.TriggerRef; ref != nil {
			result = append(result, ref.String())
		}

		return result
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build ResourceTemplate index: %w", err)
	}

	return &Reconciler{
		ReconcilerConfig: conf,
	}, nil
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	trigger := new(v1beta1.Trigger)

	if err := r.Client.Get(ctx, req.NamespacedName, trigger); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to get Trigger: %w", err)
	}

	logger := r.Logger.WithValues("trigger", trigger)
	ctx = logr.NewContext(ctx, logger)

	list := new(v1beta1.ResourceTemplateList)
	ref := &v1beta1.ObjectReference{
		APIVersion: trigger.APIVersion,
		Kind:       trigger.Kind,
		Namespace:  trigger.Namespace,
		Name:       trigger.Name,
	}
	err := r.Client.List(ctx, list, client.InNamespace(trigger.Namespace), client.MatchingFields(map[string]string{
		triggerRefField: ref.String(),
	}))
	if err != nil {
		return reconcile.Result{Requeue: true}, fmt.Errorf("failed to list ResourceTemplate: %w", err)
	}

	for _, rt := range list.Items {
		rt := rt
		result := r.patchResource(ctx, trigger, &rt)

		result.RecordEvent(r.Recorder)

		if err := result.Error; err != nil {
			logger.Error(err, result.GetMessage())

			return reconcile.Result{Requeue: result.Requeue}, err
		}

		logger.Info(result.GetMessage())
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) patchResource(ctx context.Context, trigger *v1beta1.Trigger, rt *v1beta1.ResourceTemplate) controller.Result {
	patchesBuf, err := json.Marshal(trigger.Spec.Patches)
	if err != nil {
		return controller.Result{
			Object: trigger,
			Error:  fmt.Errorf("failed to marshal patches: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	patch, err := json.Marshal([]v1beta1.JSONPatch{
		{
			Operation: v1beta1.JSONPatchOpReplace,
			Path:      "/spec/patches",
			Value:     &extv1.JSON{Raw: patchesBuf},
		},
	})
	if err != nil {
		return controller.Result{
			Object: trigger,
			Error:  fmt.Errorf("failed to marshal json patch: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	if err := r.Client.Patch(ctx, rt, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object:  trigger,
			Error:   fmt.Errorf("failed to patch resource template: %w", err),
			Reason:  ReasonPatchFailed,
			Requeue: true,
		}
	}

	return controller.Result{
		Object:  trigger,
		Message: fmt.Sprintf("Patched resource template %q", rt.Name),
		Reason:  ReasonPatched,
	}
}
