package testutil

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func MapObjects(input []runtime.Object, fn func(runtime.Object)) []runtime.Object {
	output := make([]runtime.Object, len(input))

	for i, obj := range input {
		newObj := obj.DeepCopyObject()
		fn(newObj)
		output[i] = newObj
	}

	return output
}
