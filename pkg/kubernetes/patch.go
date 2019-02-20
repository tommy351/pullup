package kubernetes

import (
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
)

var commonReducer = NewJSONPatchReducerMust(JSONPatchFromConfig([]config.ResourcePatch{
	{Remove: "/status"},
	{Replace: "/metadata/name", Value: "{{ .ModifiedName }}"},
	{Remove: "/metadata/creationTimestamp"},
	{Remove: "/metadata/resourceVersion"},
	{Remove: "/metadata/selfLink"},
	{Remove: "/metadata/uid"},
}))

var typedReducers = map[string]Reducer{
	"v1/Service": Reducers{
		NewJSONPatchReducerMust(JSONPatchFromConfig([]config.ResourcePatch{
			{Remove: "/spec/clusterIP"},
		})),
		ReducerFunc(func(data []byte, resource *Resource) ([]byte, error) {
			var service v1.Service

			if err := json.Unmarshal(data, &service); err != nil {
				return nil, err
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
		return nil, err
	}

	reducers := Reducers{commonReducer}
	obj, err := DecodeObject(raw)

	if err != nil {
		return nil, err
	}

	if r, ok := typedReducers[obj.APIVersion+"/"+obj.Kind]; ok {
		reducers = append(reducers, r)
	}

	if len(resource.Patch) > 0 {
		r, err := NewJSONPatchReducer(JSONPatchFromConfig(resource.Patch))

		if err != nil {
			return nil, err
		}

		reducers = append(reducers, r)
	}

	reducers = append(reducers, &TemplateReducer{})

	return reducers.Reduce(raw, resource)
}
