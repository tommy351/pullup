package testutil

import (
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

func ToObject(scheme *runtime.Scheme, data interface{}) (runtime.Object, error) {
	buf, err := json.Marshal(data)

	if err != nil {
		return nil, xerrors.Errorf("failed to marshal input data: %w", err)
	}

	var un unstructured.Unstructured

	if err := json.Unmarshal(buf, &un); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal data to unstructured: %w", err)
	}

	gvk := un.GetObjectKind().GroupVersionKind()
	typed, err := scheme.New(gvk)

	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			return &un, nil
		}

		return nil, xerrors.Errorf("failed to create a new object: %w", err)
	}

	if err := json.Unmarshal(buf, &typed); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal data to %s: %w", gvk, err)
	}

	typed.GetObjectKind().SetGroupVersionKind(gvk)

	return typed, nil
}

func MapObjects(input []runtime.Object, fn func(runtime.Object)) []runtime.Object {
	output := make([]runtime.Object, len(input))

	for i, obj := range input {
		newObj := obj.DeepCopyObject()
		fn(newObj)
		output[i] = newObj
	}

	return output
}
