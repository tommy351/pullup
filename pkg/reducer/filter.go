package reducer

import "golang.org/x/xerrors"

type Filter struct {
	Func func(value, key, collection interface{}) (bool, error)
}

func (f Filter) Reduce(input interface{}) (interface{}, error) {
	return reduce(input, func(e *element) (*element, error) {
		ok, err := f.Func(e.value, e.key, input)

		if err != nil {
			return nil, xerrors.Errorf("filter error: %w", err)
		}

		if !ok {
			return nil, nil
		}

		return e, nil
	})
}
