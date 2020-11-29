package resourcetemplate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates/status,verbs=get;update;patch

const (
	ReasonPatched        = "Patched"
	ReasonPatchFailed    = "PatchFailed"
	ReasonCreated        = "Created"
	ReasonCreateFailed   = "CreateFailed"
	ReasonFailed         = "Failed"
	ReasonInvalidPatch   = "InvalidPatch"
	ReasonResourceExists = "ResourceExists"
	ReasonUnchanged      = "Unchanged"
)

// ReconcilerSet provides a reconciler.
// nolint: gochecknoglobals
var ReconcilerSet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Reconciler), "*"),
)

type Logger logr.Logger

func NewLogger(logger logr.Logger) Logger {
	return logger.WithName("controller").WithName("resourcetemplate")
}

type Reconciler struct {
	Client   client.Client
	Logger   Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	rt := new(v1beta1.ResourceTemplate)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, rt); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get resource template: %w", err)
	}

	logger := r.Logger.WithValues("resourceTemplate", rt)
	ctx = logr.NewContext(ctx, logger)

	for _, patch := range rt.Spec.Patches {
		patch := patch
		result := r.applyResource(ctx, rt, &patch)
		result.RecordEvent(r.Recorder)

		if err := result.Error; err != nil {
			logger.Error(err, result.GetMessage())

			return reconcile.Result{Requeue: result.Requeue}, err
		}

		logger.Info(result.GetMessage())
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) getUnstructured(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) (*unstructured.Unstructured, error) {
	obj := new(unstructured.Unstructured)
	obj.SetGroupVersionKind(gvk)
	err := r.Client.Get(ctx, key, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get the resource: %w", err)
	}

	return obj, nil
}

func (r *Reconciler) getOriginal(ctx context.Context, gvk schema.GroupVersionKind, name types.NamespacedName) (*unstructured.Unstructured, error) {
	obj, err := r.getUnstructured(ctx, gvk, name)
	if err != nil {
		if errors.IsNotFound(err) {
			un := new(unstructured.Unstructured)
			un.SetGroupVersionKind(gvk)
			setUnstructuredNamespacedName(un, name)

			return un, nil
		}

		return nil, err
	}

	return obj, nil
}

func (r *Reconciler) applyResource(ctx context.Context, rt *v1beta1.ResourceTemplate, patch *v1beta1.WebhookPatch) controller.Result {
	gv, err := schema.ParseGroupVersion(patch.APIVersion)
	if err != nil {
		return controller.Result{
			Object: rt,
			Error:  fmt.Errorf("invalid API version: %w", err),
			Reason: ReasonInvalidPatch,
		}
	}

	patch, err = renderWebhookPatch(rt, patch)
	if err != nil {
		return controller.Result{
			Object: rt,
			Error:  err,
			Reason: ReasonInvalidPatch,
		}
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    patch.Kind,
	}

	var patchMeta strategicpatch.LookupPatchMeta
	typedObj, err := r.Scheme.New(gvk)
	if err != nil && !runtime.IsNotRegisteredError(err) {
		return controller.Result{
			Object: rt,
			Error:  fmt.Errorf("failed to create a new typed object: %w", err),
			Reason: ReasonFailed,
		}
	}

	if typedObj != nil {
		patchMeta, err = strategicpatch.NewPatchMetaFromStruct(typedObj)
		if err != nil {
			return controller.Result{
				Object: rt,
				Error:  fmt.Errorf("failed to create patch meta: %w", err),
				Reason: ReasonFailed,
			}
		}
	}

	originalName := types.NamespacedName{
		Namespace: rt.Namespace,
		Name:      patch.Name,
	}
	original, err := r.getOriginal(ctx, gvk, originalName)
	if err != nil {
		return controller.Result{
			Object:  rt,
			Error:   err,
			Reason:  ReasonFailed,
			Requeue: true,
		}
	}

	targetName := types.NamespacedName{
		Namespace: rt.Namespace,
		Name:      rt.Name,
	}
	current, err := r.getUnstructured(ctx, gvk, targetName)
	if err != nil && !errors.IsNotFound(err) {
		return controller.Result{
			Object: rt,
			Error:  fmt.Errorf("failed to get current resource: %w", err),
			Reason: ReasonFailed,
		}
	}

	if current == nil {
		desired, err := patchObjectForCreate(original, patch, patchMeta)
		if err != nil {
			return controller.Result{
				Object: rt,
				Error:  err,
				Reason: ReasonFailed,
			}
		}

		removeObjectFieldsForCreate(desired)
		setUnstructuredNamespacedName(desired, targetName)
		desired.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion:         rt.APIVersion,
				Kind:               rt.Kind,
				Name:               rt.Name,
				UID:                rt.UID,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			},
		})

		if err := r.Client.Create(ctx, desired); err != nil {
			return controller.Result{
				Object:  rt,
				Error:   fmt.Errorf("failed to create resource: %w", err),
				Reason:  ReasonCreateFailed,
				Requeue: true,
			}
		}

		return controller.Result{
			Object:  rt,
			Reason:  ReasonCreated,
			Message: fmt.Sprintf("Created resource: %s", getUnstructuredResourceName(desired)),
		}
	}

	if !metav1.IsControlledBy(current, rt) {
		return controller.Result{
			Object: rt,
			Error:  UnmanagedResourceError{Object: current},
			Reason: ReasonResourceExists,
		}
	}

	updatePatch, err := createPatchForUpdate(original, current, patch, patchMeta)
	if err != nil {
		return controller.Result{
			Object: rt,
			Error:  err,
			Reason: ReasonFailed,
		}
	}

	if updatePatch == nil {
		return controller.Result{
			Object:  rt,
			Message: fmt.Sprintf("Skipped resource: %s", getUnstructuredResourceName(current)),
			Reason:  ReasonUnchanged,
		}
	}

	if err := r.Client.Patch(ctx, current, updatePatch); err != nil {
		return controller.Result{
			Object:  rt,
			Error:   fmt.Errorf("failed to patch resource: %w", err),
			Reason:  ReasonPatchFailed,
			Requeue: true,
		}
	}

	return controller.Result{
		Object:  rt,
		Message: fmt.Sprintf("Patched resource: %s", getUnstructuredResourceName(current)),
		Reason:  ReasonPatched,
	}
}

func newJSONMap(data []byte) (strategicpatch.JSONMap, error) {
	var jsonMap strategicpatch.JSONMap

	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return jsonMap, nil
}

func getUnstructuredResourceName(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s/%s %s", obj.GetAPIVersion(), obj.GetKind(), obj.GetName())
}

func removeObjectFieldsForCreate(obj *unstructured.Unstructured) {
	unstructured.RemoveNestedField(obj.Object, "status")

	for _, key := range []string{
		"creationTimestamp",
		"resourceVersion",
		"selfLink",
		"uid",
		"generation",
		"managedFields",
	} {
		unstructured.RemoveNestedField(obj.Object, "metadata", key)
	}

	for _, key := range []string{
		"deployment.kubernetes.io/revision",
		"kubectl.kubernetes.io/last-applied-configuration",
	} {
		unstructured.RemoveNestedField(obj.Object, "metadata", "annotations", key)
	}
}

func setUnstructuredNamespacedName(obj *unstructured.Unstructured, name types.NamespacedName) {
	obj.SetNamespace(name.Namespace)
	obj.SetName(name.Name)
}
