package reducer

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrNotArrayOrMap = errors.New("expected an array or a map")

type MapFunc func(interface{}) (interface{}, error)

func MapValue(fn MapFunc) Interface {
	return MapReduceValue(Func(fn))
}

func MapReduceValue(reducer Interface) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		iv := reflect.ValueOf(input)

		switch iv.Kind() {
		case reflect.Array, reflect.Slice:
			output := reflect.MakeSlice(iv.Type(), iv.Len(), iv.Cap())

			for i := 0; i < iv.Len(); i++ {
				newValue, err := reducer.Reduce(iv.Index(i).Interface())

				if err != nil {
					return nil, fmt.Errorf("map error at index %d: %w", i, err)
				}

				output.Index(i).Set(reflect.ValueOf(newValue))
			}

			return output.Interface(), nil

		case reflect.Map:
			output := reflect.MakeMapWithSize(iv.Type(), iv.Len())
			iter := iv.MapRange()

			for iter.Next() {
				newValue, err := reducer.Reduce(iter.Value().Interface())

				if err != nil {
					return nil, fmt.Errorf("map error at key %v: %w", iter.Key().Interface(), err)
				}

				output.SetMapIndex(iter.Key(), reflect.ValueOf(newValue))
			}

			return output.Interface(), nil

		default:
			return nil, ErrNotArrayOrMap
		}
	})
}
