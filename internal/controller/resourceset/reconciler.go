package resourceset

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/reducer"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ReasonUpdated         = "Updated"
	ReasonUpdateFailed    = "UpdateFailed"
	ReasonCreated         = "Created"
	ReasonCreateFailed    = "CreateFailed"
	ReasonFailed          = "Failed"
	ReasonInvalidResource = "InvalidResource"
	ReasonResourceExists  = "ResourceExists"
	ReasonUnchanged       = "Unchanged"
)

type applyResult struct {
	Error   error
	Reason  string
	Message string
	Requeue bool
}

func (a applyResult) record(recorder record.EventRecorder, obj runtime.Object, data *unstructured.Unstructured) {
	eventType := corev1.EventTypeNormal
	msg := a.Message

	if err := a.Error; err != nil {
		eventType = corev1.EventTypeWarning

		if msg == "" {
			msg = err.Error()
		}
	}

	recorder.AnnotatedEventf(obj, map[string]string{
		"apiVersion": data.GetAPIVersion(),
		"kind":       data.GetKind(),
		"name":       data.GetName(),
	}, eventType, a.Reason, msg)
}

type Reconciler struct {
	client   client.Client
	logger   logr.Logger
	recorder record.EventRecorder
}

func NewReconciler(mgr manager.Manager, logger logr.Logger) *Reconciler {
	return &Reconciler{
		client:   mgr.GetClient(),
		logger:   logger.WithName("controller").WithName("resourceset"),
		recorder: mgr.GetEventRecorderFor("pullup"),
	}
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	set := new(v1alpha1.ResourceSet)
	ctx := context.Background()

	if err := r.client.Get(ctx, req.NamespacedName, set); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get resource set: %w", err)
	}

	set.SetGroupVersionKind(k8s.Kind("ResourceSet"))

	logger := r.logger.WithValues("resourceSet", set)
	ctx = log.NewContext(ctx, logger)

	for i, res := range set.Spec.Resources {
		var obj unstructured.Unstructured

		if err := json.Unmarshal(res, &obj); err != nil {
			r.recorder.Eventf(set, corev1.EventTypeWarning, ReasonInvalidResource, "Failed to unmarshal resource %d", i)
			return reconcile.Result{}, xerrors.Errorf("failed to unmarshal resource %d: %w", i, err)
		}

		logger := log.FromContext(ctx).WithValues("resource", obj)
		result := r.applyResource(log.NewContext(ctx, logger), set, &obj)
		result.record(r.recorder, set, &obj)

		if err := result.Error; err != nil {
			logger.Error(err, "Failed to apply resource")
			return reconcile.Result{Requeue: result.Requeue}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) getUnstructured(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) (*unstructured.Unstructured, error) {
	obj := new(unstructured.Unstructured)
	obj.SetGroupVersionKind(gvk)
	err := r.client.Get(ctx, key, obj)

	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (r *Reconciler) applyResource(ctx context.Context, set *v1alpha1.ResourceSet, obj *unstructured.Unstructured) *applyResult {
	logger := log.FromContext(ctx)
	gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())

	if err != nil {
		return &applyResult{
			Error:  xerrors.Errorf("invalid API version: %w", err),
			Reason: ReasonInvalidResource,
		}
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    obj.GetKind(),
	}

	renderedObj, err := newTemplateReducer(set).Reduce(obj.Object)

	if err != nil {
		return &applyResult{
			Error:  xerrors.Errorf("failed to render template: %w", err),
			Reason: ReasonInvalidResource,
		}
	}

	original, err := r.getUnstructured(ctx, gvk, types.NamespacedName{
		Namespace: set.Namespace,
		Name:      obj.GetName(),
	})

	if err != nil && !errors.IsNotFound(err) {
		return &applyResult{
			Error:   xerrors.Errorf("failed to get original resource: %w", err),
			Reason:  ReasonFailed,
			Requeue: true,
		}
	}

	applied, err := r.getUnstructured(ctx, gvk, types.NamespacedName{
		Namespace: set.Namespace,
		Name:      set.Name,
	})

	if err != nil && !errors.IsNotFound(err) {
		return &applyResult{
			Error:   xerrors.Errorf("failed to get last applied resource: %w", err),
			Reason:  ReasonFailed,
			Requeue: true,
		}
	}

	if applied != nil && !metav1.IsControlledBy(applied, set) {
		return &applyResult{
			Error:  xerrors.Errorf("resource already exists and is not managed by pullup: %s", getResourceName(applied)),
			Reason: ReasonResourceExists,
		}
	}

	var reducers reducer.Reducers

	if original != nil {
		reducers = append(
			reducers,
			mergeResource(original.Object),
			// Remove metadata and status from the original resource
			reducer.DeleteNested([]string{"status"}),
			reducer.ReduceNested([]string{"metadata"}, deleteKeys([]string{
				"creationTimestamp",
				"resourceVersion",
				"selfLink",
				"uid",
				"generation",
			})),
			reducer.ReduceNested([]string{"metadata", "annotations"}, deleteKeys([]string{
				"deployment.kubernetes.io/revision",
				"kubectl.kubernetes.io/last-applied-configuration",
			})),
		)

		// nolint: gocritic
		switch gvk {
		case schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}:
			// Remove cluster IP and nodePorts
			reducers = append(
				reducers,
				reducer.DeleteNested([]string{"spec", "clusterIP"}),
				reducer.ReduceNested([]string{"spec", "ports"}, reducer.MapReduceValue(deleteKeys([]string{"nodePort"}))),
			)
		}
	}

	if applied != nil {
		reducers = append(reducers, mergeResource(applied.Object))
	}

	reducers = append(
		reducers,
		mergeResource(renderedObj),
		// Set the name and owner references
		reducer.SetNested([]string{"metadata", "name"}, set.Name),
		reducer.SetNested([]string{"metadata", "namespace"}, set.Namespace),
		reducer.SetNested([]string{"metadata", "ownerReferences"}, []interface{}{
			map[string]interface{}{
				"apiVersion":         set.APIVersion,
				"kind":               set.Kind,
				"name":               set.Name,
				"uid":                set.UID,
				"controller":         true,
				"blockOwnerDeletion": true,
			},
		}),
		reducer.DeepMapValue(normalizeValue),
	)

	patch, err := reducers.Reduce(nil)

	if err != nil {
		return &applyResult{
			Error:  xerrors.Errorf("failed to reduce patches: %w", err),
			Reason: ReasonInvalidResource,
		}
	}

	logger.V(log.Debug).Info("Ready to patch the resource", "patch", patch)

	data := &unstructured.Unstructured{
		Object: patch.(map[string]interface{}),
	}

	if applied != nil {
		if equal(data, applied) {
			logger.V(log.Debug).Info("Skipped resource")
			return &applyResult{
				Reason:  ReasonUnchanged,
				Message: "Skipped resource " + getResourceName(data),
			}
		}

		if err := r.client.Update(ctx, data); err != nil {
			return &applyResult{
				Error:   xerrors.Errorf("failed to update resource: %w", err),
				Reason:  ReasonUpdateFailed,
				Requeue: true,
			}
		}

		logger.V(log.Debug).Info("Updated resource")
		return &applyResult{
			Reason:  ReasonUpdated,
			Message: "Updated resource " + getResourceName(data),
		}
	}

	if err := r.client.Create(ctx, data); err != nil {
		return &applyResult{
			Error:   xerrors.Errorf("failed to create resource: %w", err),
			Reason:  ReasonCreateFailed,
			Requeue: true,
		}
	}

	logger.V(log.Debug).Info("Created resource")

	return &applyResult{
		Reason:  ReasonCreated,
		Message: "Created resource " + getResourceName(data),
	}
}

func getResourceName(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s %s: %q", obj.GetAPIVersion(), obj.GetKind(), obj.GetName())
}
