package reducer

import (
	"reflect"

	"golang.org/x/xerrors"
)

type FilterFunc func(interface{}) (bool, error)

func FilterKey(fn FilterFunc) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		iv := reflect.ValueOf(input)

		if err := assertMapKind(iv); err != nil {
			return nil, err
		}

		output := reflect.MakeMap(iv.Type())
		iter := iv.MapRange()

		for iter.Next() {
			key := iter.Key().Interface()
			ok, err := fn(key)

			if err != nil {
				return nil, xerrors.Errorf("filter error at key %v: %w", key, err)
			}

			if ok {
				output.SetMapIndex(iter.Key(), iter.Value())
			}
		}

		return output.Interface(), nil
	})
}
