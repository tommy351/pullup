package resourcetemplate

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type UnmanagedResourceError struct {
	Object *unstructured.Unstructured
}

func (u UnmanagedResourceError) Error() string {
	return fmt.Sprintf("resource already exists and is not managed by pullup: %s", getUnstructuredResourceName(u.Object))
}
