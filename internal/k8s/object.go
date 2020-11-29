package k8s

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
