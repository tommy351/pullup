package reducer

import (
	"fmt"
)

type Interface interface {
	Reduce(input interface{}) (interface{}, error)
}

type Func func(interface{}) (interface{}, error)

func (r Func) Reduce(input interface{}) (interface{}, error) {
	return r(input)
}

type Reducers []Interface

func (r Reducers) Reduce(input interface{}) (interface{}, error) {
	var err error
	output := input

	for i, v := range r {
		if output, err = v.Reduce(output); err != nil {
			return nil, fmt.Errorf("reducer #%d returns an error: %w", i, err)
		}
	}

	return output, nil
}
