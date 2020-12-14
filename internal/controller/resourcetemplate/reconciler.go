package resourcetemplate

import (
	"context"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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
	ReasonDeleted        = "Deleted"
	ReasonDeleteFailed   = "DeleteFailed"
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
	Client    client.Client
	Logger    Logger
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	APIReader client.Reader
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	rt := new(v1beta1.ResourceTemplate)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, rt); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to get resource template: %w", err)
	}

	logger := r.Logger.WithValues("resourceTemplate", rt)

	return r.handleResourceTemplate(logr.NewContext(ctx, logger), rt)
}

func (r *Reconciler) handleResult(ctx context.Context, rt *v1beta1.ResourceTemplate, result controller.Result) (reconcile.Result, error) {
	logger := logr.FromContextOrDiscard(ctx)
	result.RecordEvent(r.Recorder, rt)

	if err := result.Error; err != nil {
		logger.Error(result.Error, result.GetMessage())
	} else {
		logger.Info(result.GetMessage())
	}

	return reconcile.Result{Requeue: result.Requeue}, result.Error
}

func (r *Reconciler) handleResourceTemplate(ctx context.Context, rt *v1beta1.ResourceTemplate) (reconcile.Result, error) {
	patches, err := r.renderTriggerPatches(ctx, rt)
	if err != nil {
		return r.handleResult(ctx, rt, controller.Result{
			Error:  err,
			Reason: ReasonInvalidPatch,
		})
	}

	activity := getResourceActivity(rt, patches)

	if err := r.updateStatus(ctx, rt, activity.Active); err != nil {
		return r.handleResult(ctx, rt, controller.Result{
			Error:   err,
			Reason:  ReasonFailed,
			Requeue: true,
		})
	}

	r.deleteInactiveResources(ctx, rt, activity.Inactive)

	for _, patch := range patches {
		patch := patch

		if result, err := r.handleResult(ctx, rt, r.applyResource(ctx, rt, &patch)); err != nil {
			return result, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) updateStatus(ctx context.Context, rt *v1beta1.ResourceTemplate, active []v1beta1.ObjectReference) error {
	now := metav1.Now()
	rt.Status.LastUpdateTime = &now
	rt.Status.Active = active

	if err := r.Client.Status().Update(ctx, rt); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (r *Reconciler) deleteInactiveResources(ctx context.Context, rt *v1beta1.ResourceTemplate, refs []v1beta1.ObjectReference) {
	for _, ref := range refs {
		obj := new(unstructured.Unstructured)
		obj.SetAPIVersion(ref.APIVersion)
		obj.SetKind(ref.Kind)
		obj.SetNamespace(ref.Namespace)
		obj.SetName(ref.Name)

		if err := r.Client.Delete(ctx, obj); err == nil {
			_, _ = r.handleResult(ctx, rt, controller.Result{
				Message: fmt.Sprintf("Deleted resource: %s", getObjectName(obj)),
				Reason:  ReasonDeleted,
			})
		} else if !errors.IsNotFound(err) {
			_, _ = r.handleResult(ctx, rt, controller.Result{
				Error:  fmt.Errorf("failed to delete resource: %w", err),
				Reason: ReasonDeleteFailed,
			})
		}
	}
}

func (r *Reconciler) getObject(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) (runtime.Object, error) {
	return k8s.GetObject(ctx, r.APIReader, r.Scheme, gvk, key)
}

func (r *Reconciler) newEmptyObject(gvk schema.GroupVersionKind, key client.ObjectKey) (runtime.Object, error) {
	obj, err := r.Scheme.New(gvk)
	if err != nil {
		if !runtime.IsNotRegisteredError(err) {
			return nil, fmt.Errorf("failed to create a new API object: %w", err)
		}

		obj = new(unstructured.Unstructured)
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)

	if err := setObjectName(obj, key); err != nil {
		return nil, err
	}

	return obj, nil
}

func (r *Reconciler) patchObject(input runtime.Object, patch *v1beta1.TriggerPatch) (runtime.Object, error) {
	inputBuf, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}

	var outputBuf []byte

	switch {
	case patch.Merge != nil && patch.Merge.Raw != nil:
		if _, ok := input.(*unstructured.Unstructured); ok {
			outputBuf, err = jsonpatch.MergePatch(inputBuf, patch.Merge.Raw)
		} else {
			outputBuf, err = strategicpatch.StrategicMergePatch(inputBuf, patch.Merge.Raw, input)
		}
	case len(patch.JSONPatch) > 0:
		outputBuf, err = applyJSONPatch(inputBuf, patch.JSONPatch)
	default:
		return input, nil
	}

	if err != nil {
		return nil, err
	}

	gvk := input.GetObjectKind().GroupVersionKind()
	accessor, err := meta.Accessor(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessor: %w", err)
	}
	name := types.NamespacedName{
		Namespace: accessor.GetNamespace(),
		Name:      accessor.GetName(),
	}

	output, err := r.newEmptyObject(gvk, name)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(outputBuf, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return output, nil
}

func (r *Reconciler) newUpdatePatch(original, desired, current runtime.Object) (client.Patch, error) {
	originalBuf, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal original resource: %w", err)
	}

	desiredBuf, err := json.Marshal(desired)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal desired resource: %w", err)
	}

	currentBuf, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current resource: %w", err)
	}

	if _, ok := current.(*unstructured.Unstructured); ok {
		return newJSONMergePatchForUpdate(originalBuf, desiredBuf, currentBuf)
	}

	return newStrategicMergePatchForUpdate(originalBuf, desiredBuf, currentBuf, current)
}

func (r *Reconciler) applyResource(ctx context.Context, rt *v1beta1.ResourceTemplate, patch *v1beta1.TriggerPatch) controller.Result {
	gvk, err := getPatchGVK(patch)
	if err != nil {
		return controller.Result{
			Error:  err,
			Reason: ReasonInvalidPatch,
		}
	}

	var original runtime.Object
	originalName := types.NamespacedName{
		Namespace: rt.Namespace,
		Name:      patch.SourceName,
	}

	if originalName.Name != "" {
		if original, err = r.getObject(ctx, gvk, originalName); err != nil && !errors.IsNotFound(err) {
			return controller.Result{
				Error:   fmt.Errorf("failed to get original resource: %w", err),
				Reason:  ReasonFailed,
				Requeue: true,
			}
		}
	}

	if original == nil {
		if original, err = r.newEmptyObject(gvk, types.NamespacedName{Namespace: rt.Namespace}); err != nil {
			return controller.Result{
				Error:  err,
				Reason: ReasonFailed,
			}
		}
	}

	desired, err := r.patchObject(original, patch)
	if err != nil {
		return controller.Result{
			Error:  err,
			Reason: ReasonInvalidPatch,
		}
	}

	currentName := types.NamespacedName{
		Namespace: rt.Namespace,
		Name:      patch.TargetName,
	}
	current, err := r.getObject(ctx, gvk, currentName)
	if err != nil && !errors.IsNotFound(err) {
		return controller.Result{
			Error:   fmt.Errorf("failed to get current resource: %w", err),
			Reason:  ReasonFailed,
			Requeue: true,
		}
	}

	if current == nil {
		obj, err := cleanObjectForCreate(desired)
		if err != nil {
			return controller.Result{
				Error:  err,
				Reason: ReasonFailed,
			}
		}

		if err := setOwnerReferences(obj, []metav1.OwnerReference{
			{
				APIVersion:         rt.APIVersion,
				Kind:               rt.Kind,
				Name:               rt.Name,
				UID:                rt.UID,
				BlockOwnerDeletion: pointer.BoolPtr(true),
				Controller:         pointer.BoolPtr(true),
			},
		}); err != nil {
			return controller.Result{
				Error:  err,
				Reason: ReasonFailed,
			}
		}

		if err := setObjectName(obj, currentName); err != nil {
			return controller.Result{
				Error:  err,
				Reason: ReasonFailed,
			}
		}

		gvk := obj.GetObjectKind().GroupVersionKind()

		if err := r.Client.Create(ctx, obj); err != nil {
			return controller.Result{
				Error:   fmt.Errorf("failed to create resource: %w", err),
				Reason:  ReasonCreateFailed,
				Requeue: true,
			}
		}

		// NOTE: The ObjectKind of obj is removed after client.Create is called.
		// It might be a bug of client package? Anyway, here I manually reset the
		// ObjectKind.
		obj.GetObjectKind().SetGroupVersionKind(gvk)

		return controller.Result{
			Message: fmt.Sprintf("Created resource: %s", getObjectName(obj)),
			Reason:  ReasonCreated,
		}
	}

	if accessor, err := meta.Accessor(current); err == nil {
		if !metav1.IsControlledBy(accessor, rt) {
			return controller.Result{
				EventType: corev1.EventTypeWarning,
				Message:   fmt.Sprintf("Resource already exists and is not managed by pullup: %s", getObjectName(current)),
				Reason:    ReasonResourceExists,
			}
		}
	}

	updatePatch, err := r.newUpdatePatch(original, desired, current)
	if err != nil {
		return controller.Result{
			Error:  err,
			Reason: ReasonFailed,
		}
	}

	if updatePatch == nil {
		return controller.Result{
			Message: fmt.Sprintf("Skipped resource: %s", getObjectName(current)),
			Reason:  ReasonUnchanged,
		}
	}

	if err := r.Client.Patch(ctx, current, updatePatch); err != nil {
		return controller.Result{
			Error:   fmt.Errorf("failed to patch resource: %w", err),
			Reason:  ReasonPatchFailed,
			Requeue: true,
		}
	}

	return controller.Result{
		Message: fmt.Sprintf("Patched resource: %s", getObjectName(current)),
		Reason:  ReasonPatched,
	}
}

func getPatchGVK(patch *v1beta1.TriggerPatch) (schema.GroupVersionKind, error) {
	gv, err := schema.ParseGroupVersion(patch.APIVersion)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("invalid API version: %w", err)
	}

	return gv.WithKind(patch.Kind), nil
}

func getObjectName(obj runtime.Object) string {
	apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	fullKind := fmt.Sprintf("%s/%s", apiVersion, kind)
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fullKind
	}

	return fmt.Sprintf("%s %s", fullKind, accessor.GetName())
}

func setObjectName(obj runtime.Object, key client.ObjectKey) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}

	accessor.SetNamespace(key.Namespace)
	accessor.SetName(key.Name)

	return nil
}

func cleanObjectForCreate(input runtime.Object) (runtime.Object, error) {
	output := input.DeepCopyObject()

	if err := cleanMetadata(output); err != nil {
		return nil, err
	}

	// nolint: gocritic
	switch obj := output.(type) {
	case *corev1.Service:
		cleanService(obj)
	}

	return output, nil
}

func cleanMetadata(obj runtime.Object) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}

	accessor.SetCreationTimestamp(metav1.Time{})
	accessor.SetResourceVersion("")
	accessor.SetSelfLink("")
	accessor.SetUID("")
	accessor.SetGeneration(0)
	accessor.SetManagedFields(nil)

	annotations := accessor.GetAnnotations()
	if annotations != nil {
		for _, key := range []string{
			"deployment.kubernetes.io/revision",
			"kubectl.kubernetes.io/last-applied-configuration",
		} {
			delete(annotations, key)
		}

		accessor.SetAnnotations(annotations)
	}

	return nil
}

func cleanService(obj *corev1.Service) {
	obj.Spec.ClusterIP = ""

	for i := range obj.Spec.Ports {
		obj.Spec.Ports[i].NodePort = 0
	}
}

func setOwnerReferences(obj runtime.Object, refs []metav1.OwnerReference) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}

	accessor.SetOwnerReferences(refs)

	return nil
}

type resourceActivity struct {
	Active   []v1beta1.ObjectReference
	Inactive []v1beta1.ObjectReference
}

func getResourceActivity(rt *v1beta1.ResourceTemplate, patches []v1beta1.TriggerPatch) *resourceActivity {
	var result resourceActivity

	inactiveMap := make(map[v1beta1.ObjectReference]struct{})

	for _, ref := range rt.Status.Active {
		inactiveMap[ref] = struct{}{}
	}

	for _, patch := range patches {
		ref := v1beta1.ObjectReference{
			APIVersion: patch.APIVersion,
			Kind:       patch.Kind,
			Namespace:  rt.Namespace,
			Name:       patch.TargetName,
		}
		result.Active = append(result.Active, ref)
		delete(inactiveMap, ref)
	}

	for ref := range inactiveMap {
		result.Inactive = append(result.Inactive, ref)
	}

	return &result
}
