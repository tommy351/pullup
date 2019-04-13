package event

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/reducer"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ResourceSetReconciler struct {
	Client client.Client
	Logger logr.Logger
}

func (r *ResourceSetReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	set := new(v1alpha1.ResourceSet)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, set); err != nil {
		return reconcile.Result{}, xerrors.Errorf("failed to get resource set: %w", err)
	}

	logger := r.Logger.WithValues("resourceSet", set)
	ctx = log.NewContext(ctx, logger)

	for _, res := range set.Spec.Resources {
		if err := r.applyResource(ctx, set, res); err != nil {
			return reconcile.Result{Requeue: true}, xerrors.Errorf("failed to apply resource: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ResourceSetReconciler) getUnstructured(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) (*unstructured.Unstructured, error) {
	obj := new(unstructured.Unstructured)
	obj.SetGroupVersionKind(gvk)
	err := r.Client.Get(ctx, key, obj)

	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (r *ResourceSetReconciler) applyResource(ctx context.Context, set *v1alpha1.ResourceSet, raw json.RawMessage) error {
	logger := log.FromContext(ctx)
	var obj unstructured.Unstructured

	if err := json.Unmarshal(raw, &obj); err != nil {
		logger.V(log.Warn).Info("Failed to unmarshal resource", log.FieldError, err) //.Err(err).Msg("Failed to unmarshal resource")
		return nil
	}

	logger = logger.WithValues("resource", raw)
	gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())

	if err != nil {
		logger.V(log.Warn).Info("Invalid API version", log.FieldError, err)
		return nil
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
		return xerrors.Errorf("failed to get original resource: %w", err)
	}

	applied, err := r.getUnstructured(ctx, gvk, types.NamespacedName{
		Namespace: set.Namespace,
		Name:      set.Name,
	})

	if err != nil && !errors.IsNotFound(err) {
		return xerrors.Errorf("failed to get applied resource: %w", err)
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
		logger.V(log.Warn).Info("Failed to render object", log.FieldError, err)
		return nil
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
		logger.V(log.Warn).Info("Failed to reduce patches", log.FieldError, err)
		return nil
	}

	logger.V(log.Debug).Info("Ready to patch the resource", "patch", patch)

	data := &unstructured.Unstructured{
		Object: patch.(map[string]interface{}),
	}

	if applied != nil {
		if err := r.Client.Update(ctx, data); err != nil {
			return xerrors.Errorf("failed to update resource: %w", err)
		}

		logger.V(log.Debug).Info("Updated resource")
		return nil
	}

	if err := r.Client.Create(ctx, data); err != nil {
		return xerrors.Errorf("failed to create resource: %w", err)
	}

	logger.V(log.Debug).Info("Created resource")
	return nil
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
