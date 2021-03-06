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

func NewEmptyObject(scheme *runtime.Scheme, gvk schema.GroupVersionKind) (client.Object, error) {
	obj, err := scheme.New(gvk)
	if err != nil {
		if !runtime.IsNotRegisteredError(err) {
			return nil, fmt.Errorf("failed to create a new API object: %w", err)
		}

		un := new(unstructured.Unstructured)
		un.SetGroupVersionKind(gvk)

		return un, nil
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)

	return obj.(client.Object), nil
}

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

func ToObject(scheme *runtime.Scheme, data interface{}) (client.Object, error) {
	if obj, ok := data.(client.Object); ok {
		return obj, nil
	}

	un, err := ToUnstructured(data)
	if err != nil {
		return nil, err
	}

	gvk := un.GetObjectKind().GroupVersionKind()
	typed, err := NewEmptyObject(scheme, gvk)
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

func MapObjects(input []client.Object, fn func(client.Object) error) ([]client.Object, error) {
	output := make([]client.Object, len(input))

	for i, obj := range input {
		newObj := obj.DeepCopyObject().(client.Object)

		if err := fn(newObj); err != nil {
			return nil, err
		}

		output[i] = newObj
	}

	return output, nil
}

func GetObject(ctx context.Context, reader client.Reader, scheme *runtime.Scheme, gvk schema.GroupVersionKind, key client.ObjectKey) (client.Object, error) {
	obj, err := NewEmptyObject(scheme, gvk)
	if err != nil {
		return nil, err
	}

	// Use APIReader to disable the cache
	if err := reader.Get(ctx, key, obj); err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)

	return obj, nil
}
