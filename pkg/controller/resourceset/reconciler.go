package resourceset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/reducer"
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
	set := new(v1alpha1.ResourceSet)
	ctx := context.Background()

	if err := r.client.Get(ctx, req.NamespacedName, set); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get resource set: %w", err)
	}

	logger := r.logger.WithValues("resourceSet", set)
	ctx = log.NewContext(ctx, logger)

	for i, res := range set.Spec.Resources {
		var obj unstructured.Unstructured

		if err := json.Unmarshal(res, &obj); err != nil {
			r.EventRecorder.Eventf(set, corev1.EventTypeWarning, ReasonInvalidResource, "Failed to unmarshal resource %d", i)
			return reconcile.Result{}, xerrors.Errorf("failed to unmarshal resource %d: %w", i, err)
		}

		logger := log.FromContext(ctx).WithValues("resource", obj)
		result := r.applyResource(log.NewContext(ctx, logger), set, &obj)
		result.record(r.EventRecorder, set, &obj)

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
			Error:  xerrors.New("resource already exists and is not managed by pullup"),
			Reason: ReasonResourceExists,
		}
	}

	var reducers []reducer.Interface

	if original != nil {
		reducers = append(
			reducers,
			reducer.Merge{Source: original.Object},
			// Remove metadata and status from the original resource
			reducer.Merge{Source: map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": nil,
					"resourceVersion":   nil,
					"selfLink":          nil,
					"uid":               nil,
					"generation":        nil,
					"annotations": map[string]interface{}{
						"deployment.kubernetes.io/revision": nil,
					},
				},
				"status": nil,
			}},
		)

		// nolint: gocritic
		switch gvk {
		case schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}:
			// Remove cluster IP and nodePorts
			reducers = append(reducers, reducer.Merge{Source: map[string]interface{}{
				"spec": map[string]interface{}{
					"clusterIP": nil,
					"ports": reducer.Map{Func: func(value, _, _ interface{}) (interface{}, error) {
						return reducer.Pipe(value, reducer.Filter{Func: func(_, key, _ interface{}) (bool, error) {
							return key != "nodePort", nil
						}})
					}},
				},
			}})
		}
	}

	if applied != nil {
		reducers = append(reducers, reducer.Merge{Source: applied.Object})
	}

	renderedObj, err := newTemplateReducer(set).Reduce(obj.Object)

	if err != nil {
		return &applyResult{
			Error:  xerrors.Errorf("failed to render template: %w", err),
			Reason: ReasonInvalidResource,
		}
	}

	reducers = append(
		reducers,
		reducer.Merge{Source: renderedObj},
		// Set the name and owner references
		reducer.Merge{Source: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": set.Name,
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion":         set.APIVersion,
						"kind":               set.Kind,
						"name":               set.Name,
						"uid":                set.UID,
						"controller":         true,
						"blockOwnerDeletion": true,
					},
				},
			},
		}},
	)

	patch, err := reducer.Pipe(nil, reducers...)

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
			Message: "Updated resource " + getResourceName(obj),
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
		Message: "Created resource " + getResourceName(obj),
	}
}

func getResourceName(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s %s: %q", obj.GetAPIVersion(), obj.GetKind(), obj.GetName())
}

func newTemplateReducer(data interface{}) reducer.Interface {
	var mapper reducer.Map

	mapper = reducer.Map{Func: func(value, key, _ interface{}) (interface{}, error) {
		v, err := mapper.Reduce(value)

		if err != nil {
			if !xerrors.Is(err, reducer.ErrNotCollection) {
				return nil, xerrors.Errorf("render error at key %v: %w", key, err)
			}

			if s, ok := value.(string); ok {
				tmpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(s)

				if err != nil {
					return nil, xerrors.Errorf("failed to parse template: %w", err)
				}

				var buf bytes.Buffer

				if err := tmpl.Execute(&buf, data); err != nil {
					return nil, xerrors.Errorf("failed to execute template: %w", err)
				}

				return buf.String(), nil
			}

			return value, nil
		}

		return v, nil
	}}

	return mapper
}
