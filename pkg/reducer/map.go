package reducer

import "golang.org/x/xerrors"

type Map struct {
	Func func(value, key, collection interface{}) (interface{}, error)
}

func (m Map) Reduce(input interface{}) (interface{}, error) {
	return reduce(input, func(e *element) (*element, error) {
		v, err := m.Func(e.value, e.key, input)

		if err != nil {
			return nil, xerrors.Errorf("map error: %w", err)
		}

		return &element{key: e.key, value: v}, nil
	})
}
