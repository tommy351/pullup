package kubernetes

import (
	"github.com/ansel1/merry"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
)

type Reducer interface {
	Reduce(resource *Resource) error
}

type Reducers []Reducer

func (r Reducers) Reduce(resource *Resource) error {
	for _, reducer := range r {
		if err := reducer.Reduce(resource); err != nil {
			return merry.Wrap(err)
		}
	}

	return nil
}

type ReducerFunc func(resource *Resource) error

func (r ReducerFunc) Reduce(resource *Resource) error {
	return r(resource)
}

func newUnstructuredFromRawJSON(buf []byte) (*unstructured.Unstructured, error) {
	var data map[string]interface{}

	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, merry.Wrap(err)
	}

	return &unstructured.Unstructured{Object: data}, nil
}

type ByteReducerFunc func(data []byte, resource *Resource) ([]byte, error)

func (b ByteReducerFunc) Reduce(resource *Resource) error {
	buf, err := resource.PatchedResource.MarshalJSON()

	if err != nil {
		return merry.Wrap(err)
	}

	buf, err = b(buf, resource)

	if err != nil {
		return merry.Wrap(err)
	}

	resource.PatchedResource, err = newUnstructuredFromRawJSON(buf)
	return merry.Wrap(err)
}

type ConditionalReducer struct {
	condition func(resource *Resource) bool
	reducer   Reducer
}

func NewConditionalReducer(condition func(resource *Resource) bool, reducer Reducer) *ConditionalReducer {
	return &ConditionalReducer{
		condition: condition,
		reducer:   reducer,
	}
}

func (c ConditionalReducer) Reduce(resource *Resource) error {
	if c.condition(resource) {
		return c.reducer.Reduce(resource)
	}

	return nil
}
