package testenv

import (
	"context"
	"fmt"
	"sort"

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
	// nolint: prealloc
	var (
		recorders []Recorder
		changes   []Change
	)

	appendRecorder := func(value interface{}) {
		if v, ok := value.(Recorder); ok {
			recorders = append(recorders, v)
		}
	}

	switch c := c.(type) {
	case Recorder:
		appendRecorder(c)

	case client.DelegatingClient:
		appendRecorder(c.Writer)
		appendRecorder(c.StatusClient)

	case *client.DelegatingClient:
		appendRecorder(c.Writer)
		appendRecorder(c.StatusClient)
	}

	for _, r := range recorders {
		changes = append(changes, r.Changes()...)
	}

	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].NamespacedName.String() < changes[j].NamespacedName.String()
	})

	return changes
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
				return nil, fmt.Errorf("failed to create a new object: %w", err)
			}

			obj = new(unstructured.Unstructured)
			obj.GetObjectKind().SetGroupVersionKind(event.GroupVersionKind)
		}

		if err := client.Get(ctx, event.NamespacedName, obj); err != nil {
			return nil, fmt.Errorf("failed to get the object: %w", err)
		}

		obj.GetObjectKind().SetGroupVersionKind(event.GroupVersionKind)
		objects[i] = obj
	}

	return objects, nil
}
