package reducer

import (
	"encoding/json"

	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ReduceBytes(resource *model.Resource, reducer func(data []byte, resource *model.Resource) ([]byte, error)) error {
	buf, err := resource.PatchedResource.MarshalJSON()

	if err != nil {
		return merry.Wrap(err)
	}

	buf, err = reducer(buf, resource)

	if err != nil {
		return merry.Wrap(err)
	}

	resource.PatchedResource, err = newUnstructuredFromRawJSON(buf)
	return merry.Wrap(err)
}

func newUnstructuredFromRawJSON(buf []byte) (*unstructured.Unstructured, error) {
	var data map[string]interface{}

	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, merry.Wrap(err)
	}

	return &unstructured.Unstructured{Object: data}, nil
}
