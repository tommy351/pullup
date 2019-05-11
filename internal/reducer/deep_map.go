package reducer

import (
	"golang.org/x/xerrors"
)

func DeepMapValue(fn MapFunc) Interface {
	var reducer Interface

	reducer = MapValue(func(value interface{}) (interface{}, error) {
		newValue, err := fn(value)

		if err != nil {
			return nil, err
		}

		reducedValue, err := reducer.Reduce(newValue)

		if err != nil {
			if xerrors.Is(err, ErrNotArrayOrMap) {
				return newValue, nil
			}

			return nil, xerrors.Errorf("deep map error: %w", err)
		}

		return reducedValue, nil
	})

	return reducer
}
