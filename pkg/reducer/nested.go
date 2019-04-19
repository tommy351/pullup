package reducer

import (
	"reflect"

	"golang.org/x/xerrors"
)

var ErrNotMap = xerrors.New("expected a map")

func assertMapKind(v reflect.Value) error {
	if v.Kind() == reflect.Map {
		return nil
	}

	return ErrNotMap
}

func ReduceNested(keys []string, reducer Interface) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		output := DeepCopy(input)
		value := output

		for i, k := range keys {
			key := reflect.ValueOf(k)
			v := reflect.ValueOf(value)

			if err := assertMapKind(v); err != nil {
				return nil, err
			}

			if i == len(keys)-1 {
				newValue, err := reducer.Reduce(v.MapIndex(key).Interface())

				if err != nil {
					return nil, xerrors.Errorf("reduce error: %w", err)
				}

				v.SetMapIndex(key, reflect.ValueOf(newValue))
			} else {
				value = v.MapIndex(key).Interface()
			}
		}

		return output, nil
	})
}

func SetNested(keys []string, value interface{}) Interface {
	return ReduceNested(keys, Func(func(_ interface{}) (interface{}, error) {
		return value, nil
	}))
}

func DeleteNested(keys []string) Interface {
	lastKey := keys[len(keys)-1]

	return ReduceNested(keys[0:len(keys)-1], FilterKey(func(key interface{}) (bool, error) {
		return key != lastKey, nil
	}))
}
