package reducer

import (
	"reflect"
	"strings"

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
		value := reflect.ValueOf(output)
		parent := value

		for _, k := range keys {
			if err := assertMapKind(value); err != nil {
				return nil, err
			}

			key := reflect.ValueOf(k)
			parent = value
			value = value.MapIndex(key)

			if value.Kind() == reflect.Interface {
				value = value.Elem()
			}
		}

		var orig interface{}

		if value.IsValid() {
			orig = value.Interface()
		}

		newValue, err := reducer.Reduce(orig)

		if err != nil {
			return nil, xerrors.Errorf("reduce error at key %s: %w", strings.Join(keys, "."), err)
		}

		parent.SetMapIndex(reflect.ValueOf(keys[len(keys)-1]), reflect.ValueOf(newValue))
		return output, nil
	})
}

func SetNested(keys []string, value interface{}) Interface {
	return ReduceNested(keys, Func(func(_ interface{}) (interface{}, error) {
		return value, nil
	}))
}

func DeleteNested(keys []string) Interface {
	lastIndex := len(keys) - 1
	lastKey := keys[lastIndex]
	headKeys := keys[0:lastIndex]

	if len(headKeys) == 0 {
		return FilterKey(func(key interface{}) (bool, error) {
			return key != lastKey, nil
		})
	}

	return ReduceNested(headKeys, FilterKey(func(key interface{}) (bool, error) {
		return key != lastKey, nil
	}))
}
