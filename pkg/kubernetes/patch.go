package kubernetes

import (
	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
)

// nolint: gochecknoglobals
var commonReducer = MustNewJSONPatchReducer(JSONPatchFromConfig([]config.ResourcePatch{
	{Remove: "/status"},
	{Replace: "/metadata/name", Value: "{{ .ModifiedName }}"},
	{Remove: "/metadata/creationTimestamp"},
	{Remove: "/metadata/resourceVersion"},
	{Remove: "/metadata/selfLink"},
	{Remove: "/metadata/uid"},
}))

// nolint: gochecknoglobals
var typedReducers = map[string]Reducer{
	"v1/Service": Reducers{
		MustNewJSONPatchReducer(JSONPatchFromConfig([]config.ResourcePatch{
			{Remove: "/spec/clusterIP"},
		})),
		ReducerFunc(func(data []byte, resource *Resource) ([]byte, error) {
			var service corev1.Service

			if err := json.Unmarshal(data, &service); err != nil {
				return nil, merry.Wrap(err)
			}

			for i := range service.Spec.Ports {
				service.Spec.Ports[i].NodePort = 0
			}

			return json.Marshal(&service)
		}),
	},
}

func PatchResource(input rest.Result, resource *Resource) ([]byte, error) {
	raw, err := input.Raw()

	if err != nil {
		return nil, merry.Wrap(err)
	}

	reducers := Reducers{commonReducer}

	obj, err := DecodeObject(raw)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	if r, ok := typedReducers[obj.APIVersion+"/"+obj.Kind]; ok {
		reducers = append(reducers, r)
	}

	if len(resource.Patch) > 0 {
		r, err := NewJSONPatchReducer(JSONPatchFromConfig(resource.Patch))

		if err != nil {
			return nil, merry.Wrap(err)
		}

		reducers = append(reducers, r)
	}

	reducers = append(reducers, &TemplateReducer{})
	result, err := reducers.Reduce(raw, resource)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return result, nil
}
