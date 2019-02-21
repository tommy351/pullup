package kubernetes

import (
	"fmt"

	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// nolint: gochecknoglobals
var metadataFieldsToRemove = []string{"creationTimestamp", "resourceVersion", "selfLink", "uid"}

// nolint: gochecknoglobals
var commonReducer = ReducerFunc(func(resource *Resource) error {
	applied := resource.AppliedResource
	patched := resource.PatchedResource

	if applied == nil {
		// Set name
		patched.SetName(resource.ModifiedName())

		// Remove metadata
		for _, field := range metadataFieldsToRemove {
			unstructured.RemoveNestedField(patched.Object, "metadata", field)
		}

		return nil
	}

	// Copy metadata from the applied resource
	meta, ok, err := unstructured.NestedMap(applied.Object, "metadata")

	if !ok {
		return merry.Wrap(err)
	}

	return unstructured.SetNestedMap(patched.Object, meta, "metadata")
})

// nolint: gochecknoglobals
var typedReducers = map[string]Reducer{
	"v1/Service": ReducerFunc(func(resource *Resource) error {
		patched := resource.PatchedResource

		if applied := resource.AppliedResource; applied != nil {
			// Copy spec from the applied resource
			spec, ok, err := unstructured.NestedMap(applied.Object, "spec")

			if !ok {
				return merry.Wrap(err)
			}

			return unstructured.SetNestedMap(patched.Object, spec, "spec")
		}

		// Remove spec.clusterIP
		unstructured.RemoveNestedField(patched.Object, "spec", "clusterIP")

		// Remove spec.ports.*.nodePort
		var newPorts []interface{}
		ports, ok, err := unstructured.NestedSlice(patched.Object, "spec", "ports")

		if !ok {
			return merry.Wrap(err)
		}

		for _, port := range ports {
			portMap := port.(map[string]interface{})
			delete(portMap, "nodePort")
			newPorts = append(newPorts, portMap)
		}

		return unstructured.SetNestedSlice(patched.Object, newPorts, "spec", "ports")
	}),
}

type Resource struct {
	config.ResourceConfig

	PullRequestNumber int
	HeadCommitSHA     string
	OriginalResource  *unstructured.Unstructured
	AppliedResource   *unstructured.Unstructured
	PatchedResource   *unstructured.Unstructured
}

func (r *Resource) ModifiedName() string {
	return fmt.Sprintf("%s-pullup-%d", r.Name, r.PullRequestNumber)
}

func (r *Resource) Patch() error {
	reducers := Reducers{commonReducer}

	if r, ok := typedReducers[r.OriginalResource.GetAPIVersion()+"/"+r.OriginalResource.GetKind()]; ok {
		reducers = append(reducers, r)
	}

	if patches := r.ResourceConfig.Patch; len(patches) > 0 {
		r, err := NewJSONPatchReducer(JSONPatchFromConfig(patches))

		if err != nil {
			return merry.Wrap(err)
		}

		reducers = append(reducers, r)
	}

	reducers = append(reducers, &TemplateReducer{})
	r.PatchedResource = r.OriginalResource.DeepCopy()

	return merry.Wrap(reducers.Reduce(r))
}
