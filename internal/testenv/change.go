package testenv

import (
	"context"
	"fmt"
	"sort"

	"github.com/tommy351/pullup/internal/k8s"
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

	case DelegatingClient:
		appendRecorder(c.Writer)
		appendRecorder(c.StatusClient)

	case *DelegatingClient:
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

func GetChangedObjects(changes []Change) ([]client.Object, error) {
	objects := make([]client.Object, len(changes))
	ctx := context.Background()
	client := GetClient()
	scheme := GetScheme()

	for i, event := range changes {
		obj, err := k8s.GetObject(ctx, client, scheme, event.GroupVersionKind, event.NamespacedName)
		if err != nil {
			return nil, fmt.Errorf("failed to get the object: %w", err)
		}

		objects[i] = obj
	}

	return objects, nil
}
