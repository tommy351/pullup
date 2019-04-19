package reducer

import (
	"reflect"

	"golang.org/x/xerrors"
)

var ErrNotArrayOrMap = xerrors.New("expected an array or a map")

type MapFunc func(interface{}) (interface{}, error)

func MapValue(fn MapFunc) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		iv := reflect.ValueOf(input)

		switch iv.Kind() {
		case reflect.Array, reflect.Slice:
			output := reflect.MakeSlice(iv.Type(), iv.Cap(), iv.Len())

			for i := 0; i < iv.Len(); i++ {
				newValue, err := fn(iv.Index(i).Interface())

				if err != nil {
					return nil, xerrors.Errorf("map error at index %d: %w", i, err)
				}

				output.Index(i).Set(reflect.ValueOf(newValue))
			}

			return output.Interface(), nil

		case reflect.Map:
			output := reflect.MakeMapWithSize(iv.Type(), iv.Len())
			iter := iv.MapRange()

			for iter.Next() {
				newValue, err := fn(iter.Value().Interface())

				if err != nil {
					return nil, xerrors.Errorf("map error at key %v: %w", iter.Key().Interface(), err)
				}

				output.SetMapIndex(iter.Key(), reflect.ValueOf(newValue))
			}

			return output.Interface(), nil

		default:
			return nil, ErrNotArrayOrMap
		}
	})
}
