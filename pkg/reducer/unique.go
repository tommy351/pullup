package reducer

import "golang.org/x/xerrors"

type Unique struct {
	Func func(value, key, collection interface{}) (interface{}, error)
}

func (u Unique) Reduce(input interface{}) (interface{}, error) {
	keys := map[interface{}]struct{}{}

	filter := Filter{
		Func: func(value, key, collection interface{}) (bool, error) {
			k, err := u.Func(value, key, collection)

			if err != nil {
				return false, xerrors.Errorf("unique error: %w", err)
			}

			_, ok := keys[k]

			if !ok {
				keys[k] = struct{}{}
			}

			return !ok, nil
		},
	}

	return filter.Reduce(input)
}
