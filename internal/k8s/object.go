package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ToUnstructured(data interface{}) (*unstructured.Unstructured, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %w", err)
	}

	var un unstructured.Unstructured

	if err := json.Unmarshal(buf, &un); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data to unstructured: %w", err)
	}

	return &un, nil
}

func ToObject(scheme *runtime.Scheme, data interface{}) (runtime.Object, error) {
	if obj, ok := data.(runtime.Object); ok {
		return obj, nil
	}

	un, err := ToUnstructured(data)
	if err != nil {
		return nil, err
	}

	gvk := un.GetObjectKind().GroupVersionKind()
	typed, err := scheme.New(gvk)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			return un, nil
		}

		return nil, fmt.Errorf("failed to create a new object: %w", err)
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %w", err)
	}

	if err := json.Unmarshal(buf, &typed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data to %s: %w", gvk, err)
	}

	typed.GetObjectKind().SetGroupVersionKind(gvk)

	return typed, nil
}

func MapObjects(input []runtime.Object, fn func(runtime.Object) error) ([]runtime.Object, error) {
	output := make([]runtime.Object, len(input))

	for i, obj := range input {
		newObj := obj.DeepCopyObject()

		if err := fn(newObj); err != nil {
			return nil, err
		}

		output[i] = newObj
	}

	return output, nil
}

func GetObject(ctx context.Context, reader client.Reader, scheme *runtime.Scheme, gvk schema.GroupVersionKind, key client.ObjectKey) (runtime.Object, error) {
	obj, err := scheme.New(gvk)
	if err != nil {
		if !runtime.IsNotRegisteredError(err) {
			return nil, fmt.Errorf("failed to create a new API object: %w", err)
		}

		un := new(unstructured.Unstructured)
		un.SetGroupVersionKind(gvk)
		obj = un
	}

	// Use APIReader to disable the cache
	if err := reader.Get(ctx, key, obj); err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)

	return obj, nil
}
