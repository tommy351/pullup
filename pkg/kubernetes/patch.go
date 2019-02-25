package kubernetes

import (
	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/model"
	"github.com/tommy351/pullup/pkg/reducer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// nolint: gochecknoglobals
var commonReducer = &reducer.If{
	Condition: func(resource *model.Resource) bool {
		return resource.AppliedResource == nil
	},
	True: reducer.Must(reducer.NewJSONPatch(reducer.JSONPatchFromConfig([]config.ResourcePatch{
		{Replace: "/metadata/name", Value: "{{ .ModifiedName }}"},
		{Remove: "/metadata/creationTimestamp"},
		{Remove: "/metadata/resourceVersion"},
		{Remove: "/metadata/selfLink"},
		{Remove: "/metadata/uid"},
	}))),
	False: reducer.Func(func(resource *model.Resource) error {
		return copyNestedField(resource.AppliedResource, resource.PatchedResource, "metadata")
	}),
}

// nolint: gochecknoglobals
var typedReducers = map[string]reducer.Reducer{
	"v1/Service": reducer.Func(func(resource *model.Resource) error {
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

func Patch(resource *model.Resource) error {
	reducers := reducer.List{commonReducer}

	if r, ok := typedReducers[resource.OriginalResource.GetAPIVersion()+"/"+resource.OriginalResource.GetKind()]; ok {
		reducers = append(reducers, r)
	}

	if patches := resource.ResourceConfig.Patch; len(patches) > 0 {
		r, err := reducer.NewJSONPatch(reducer.JSONPatchFromConfig(patches))

		if err != nil {
			return merry.Wrap(err)
		}

		reducers = append(reducers, r)
	}

	reducers = append(reducers, &reducer.Template{})
	resource.PatchedResource = resource.OriginalResource.DeepCopy()

	return merry.Wrap(reducers.Reduce(resource))
}
