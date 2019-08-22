package testenv

import (
	"context"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Change struct {
	schema.GroupVersionKind
	types.NamespacedName

	Type string
}

func GetChanges(c interface{}) []Change {
	var recorder Recorder

	switch c := c.(type) {
	case Recorder:
		recorder = c
	case client.DelegatingClient:
		recorder = c.Writer.(Recorder)
	case *client.DelegatingClient:
		recorder = c.Writer.(Recorder)
	}

	return recorder.Changes()
}

func GetChangedObjects(changes []Change) ([]runtime.Object, error) {
	objects := make([]runtime.Object, len(changes))
	ctx := context.Background()
	client := GetClient()
	scheme := GetScheme()

	for i, event := range changes {
		obj, err := scheme.New(event.GroupVersionKind)

		if err != nil {
			if !runtime.IsNotRegisteredError(err) {
				return nil, xerrors.Errorf("failed to create a new object: %w", err)
			}

			obj = new(unstructured.Unstructured)
			obj.GetObjectKind().SetGroupVersionKind(event.GroupVersionKind)
		}

		if err := client.Get(ctx, event.NamespacedName, obj); err != nil {
			return nil, xerrors.Errorf("failed to get the object: %w", err)
		}

		obj.GetObjectKind().SetGroupVersionKind(event.GroupVersionKind)
		objects[i] = obj
	}

	return objects, nil
}