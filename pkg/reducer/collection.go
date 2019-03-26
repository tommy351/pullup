package reducer

import (
	"reflect"

	"golang.org/x/xerrors"
)

var ErrNotCollection = xerrors.New("expected an array or a map")

type element struct {
	key   interface{}
	value interface{}
}

func reduce(input interface{}, fn func(e *element) (*element, error)) (interface{}, error) {
	v := reflect.ValueOf(input)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		var acc []interface{}

		for i := 0; i < v.Len(); i++ {
			output, err := fn(&element{key: i, value: v.Index(i).Interface()})

			if err != nil {
				return nil, xerrors.Errorf("reduce error at index %d: %w", i, err)
			}

			if output != nil {
				acc = append(acc, output.value)
			}
		}

		return acc, nil

	case reflect.Map:
		acc := map[string]interface{}{}

		for _, key := range v.MapKeys() {
			output, err := fn(&element{key: key.Interface(), value: v.MapIndex(key).Interface()})

			if err != nil {
				return nil, xerrors.Errorf("reduce error at key %v: %w", key.Interface(), err)
			}

			if output != nil {
				key, ok := output.key.(string)

				if !ok {
					return nil, xerrors.New("key must be a string")
				}

				acc[key] = output.value
			}
		}

		return acc, nil

	default:
		return nil, ErrNotCollection
	}
}
