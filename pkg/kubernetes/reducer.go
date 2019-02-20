package kubernetes

import "github.com/ansel1/merry"

type Reducer interface {
	Reduce(data []byte, resource *Resource) ([]byte, error)
}

type Reducers []Reducer

func (r Reducers) Reduce(input []byte, resource *Resource) (output []byte, err error) {
	output = input

	for _, reducer := range r {
		if output, err = reducer.Reduce(output, resource); err != nil {
			return nil, merry.Wrap(err)
		}
	}

	return output, nil
}

type ReducerFunc func(data []byte, resource *Resource) ([]byte, error)

func (r ReducerFunc) Reduce(data []byte, resource *Resource) ([]byte, error) {
	return r(data, resource)
}
