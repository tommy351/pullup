package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"strconv"

	"github.com/Masterminds/sprig"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceSetEventHandler struct {
	Client *k8s.Client
}

func (r *ResourceSetEventHandler) OnUpdate(ctx context.Context, obj interface{}) error {
	set, ok := obj.(*v1alpha1.ResourceSet)

	if !ok {
		return nil
	}

	for _, res := range set.Spec.Resources {
		if err := r.applyResource(ctx, set, res); err != nil {
			return xerrors.Errorf("failed to apply resource: %w", err)
		}
	}

	return nil
}

func (*ResourceSetEventHandler) OnDelete(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *ResourceSetEventHandler) applyResource(ctx context.Context, set *v1alpha1.ResourceSet, raw json.RawMessage) error {
	var obj unstructured.Unstructured

	if err := json.Unmarshal(raw, &obj); err != nil {
		return xerrors.Errorf("failed to unmarshal resource: %w", err)
	}

	logger := zerolog.Ctx(ctx).With().
		Str("resourceSet", set.Name).
		Str("apiVersion", obj.GetAPIVersion()).
		Str("kind", obj.GetKind()).
		Str("name", obj.GetName()).
		Logger()

	gvr, err := k8s.ParseGVR(obj.GetAPIVersion(), obj.GetKind())

	if err != nil {
		logger.Warn().Err(err).Msg("Invalid API version")
		return nil
	}

	client := r.Client.NewDynamicInterface(ctx, gvr)
	original, err := client.Get(obj.GetName(), metav1.GetOptions{})

	if err != nil && !k8s.IsNotFoundError(err) {
		return xerrors.Errorf("failed to get original resource: %w", err)
	}

	applied, err := client.Get(set.Name, metav1.GetOptions{})

	if err != nil && !k8s.IsNotFoundError(err) {
		return xerrors.Errorf("failed to get applied resource: %w", err)
	}

	var patch interface{}

	if original != nil {
		patch = mergeValue(patch, original.Object)
	}

	// Remove metadata and status from the original resource
	patch = mergeValue(patch, map[string]interface{}{
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
	})

	// nolint: gocritic
	switch gvr {
	case schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}:
		// Remove cluster IP
		patch = mergeValue(patch, map[string]interface{}{
			"spec": map[string]interface{}{
				"clusterIP": nil,
			},
		})
	}

	if applied != nil {
		patch = mergeValue(patch, applied.Object)
	}

	patch = mergeValue(patch, obj.Object)

	// Set the name and owner references
	patch = mergeValue(patch, map[string]interface{}{
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
	})

	patch, err = deepRenderTemplate(patch, set, "")

	if err != nil {
		logger.Warn().Err(err).Msg("Failed to render template")
		return nil
	}

	data := &unstructured.Unstructured{
		Object: patch.(map[string]interface{}),
	}

	if applied != nil {
		if _, err := client.Update(data, metav1.UpdateOptions{}); err != nil {
			return xerrors.Errorf("failed to update resource: %w", err)
		}

		logger.Debug().Msg("Updated resource")
		return nil
	}

	if _, err := client.Create(data, metav1.CreateOptions{}); err != nil {
		return xerrors.Errorf("failed to create resource: %w", err)
	}

	logger.Debug().Msg("Created resource")
	return nil
}

func mergeValue(base, patch interface{}) interface{} {
	switch patch := patch.(type) {
	case []interface{}:
		baseArr, ok := base.([]interface{})

		if !ok {
			return patch
		}

		var newArr []interface{}

		copy(newArr, baseArr)

		for i, v := range patch {
			newArr[i] = mergeValue(newArr[i], v)
		}

		return newArr

	case map[string]interface{}:
		baseMap, ok := base.(map[string]interface{})

		if !ok {
			return patch
		}

		newMap := map[string]interface{}{}

		for k, v := range baseMap {
			newMap[k] = v
		}

		for k, v := range patch {
			newMap[k] = mergeValue(newMap[k], v)
		}

		return newMap

	default:
		return patch
	}
}

func deepRenderTemplate(input interface{}, data interface{}, parent string) (interface{}, error) {
	switch input := input.(type) {
	case string:
		tmpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(input)

		if err != nil {
			return nil, xerrors.Errorf("failed to parse template: %w", err)
		}

		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, xerrors.Errorf("failed to execute template: %w", err)
		}

		return buf.String(), nil

	case []interface{}:
		var err error
		output := make([]interface{}, len(input))

		for i, v := range input {
			path := parent + strconv.Itoa(i)

			if output[i], err = deepRenderTemplate(v, data, path+"."); err != nil {
				return nil, xerrors.Errorf("render failed at path %q: %w", path, err)
			}
		}

		return output, nil

	case map[string]interface{}:
		var err error
		output := make(map[string]interface{}, len(input))

		for k, v := range input {
			path := parent + k

			if output[k], err = deepRenderTemplate(v, data, path+"."); err != nil {
				return nil, xerrors.Errorf("render failed at path %q: %w", path, err)
			}
		}

		return output, nil
	}

	return input, nil
}
