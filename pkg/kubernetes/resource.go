package kubernetes

import (
	"fmt"

	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// nolint: gochecknoglobals
var commonReducer = Reducers{
	NewConditionalReducer(func(resource *Resource) bool {
		return resource.AppliedResource == nil
	}, MustNewJSONPatchReducer(JSONPatchFromConfig([]config.ResourcePatch{
		{Replace: "/metadata/name", Value: "{{ .ModifiedName }}"},
		{Remove: "/metadata/creationTimestamp"},
		{Remove: "/metadata/resourceVersion"},
		{Remove: "/metadata/selfLink"},
		{Remove: "/metadata/uid"},
	}))),
	NewConditionalReducer(func(resource *Resource) bool {
		return resource.AppliedResource != nil
	}, ReducerFunc(func(resource *Resource) error {
		return copyNestedField(resource.AppliedResource, resource.PatchedResource, "metadata")
	})),
}

// nolint: gochecknoglobals
var typedReducers = map[string]Reducer{
	"v1/Service": ReducerFunc(func(resource *Resource) error {
		patched := resource.PatchedResource

		if applied := resource.AppliedResource; applied != nil {
			return copyNestedField(applied, patched, "spec")
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

		return merry.Wrap(unstructured.SetNestedSlice(patched.Object, newPorts, "spec", "ports"))
	}),
}

func copyNestedField(src, dst *unstructured.Unstructured, fields ...string) error {
	value, ok, err := unstructured.NestedFieldCopy(src.Object, fields...)

	if !ok {
		return merry.Wrap(err)
	}

	return merry.Wrap(unstructured.SetNestedField(dst.Object, value, fields...))
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
