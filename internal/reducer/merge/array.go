package merge

import (
	"fmt"
	"reflect"

	"k8s.io/utils/integer"
)

func (m *Merger) mergeArray(input, source reflect.Value) (interface{}, error) {
	switch source.Kind() {
	case reflect.Array, reflect.Slice:
		size := integer.IntMax(input.Len(), source.Len())
		arrayType := reflect.SliceOf(getUnionType(getArrayElementType(input), getArrayElementType(source)))
		output := reflect.MakeSlice(arrayType, size, size)

		// Set input values
		for i := 0; i < input.Len(); i++ {
			output.Index(i).Set(input.Index(i))
		}

		// Merge source values
		for i := 0; i < source.Len(); i++ {
			if i < input.Len() {
				newValue, err := m.Func(input.Index(i).Interface(), source.Index(i).Interface())
				if err != nil {
					return nil, fmt.Errorf("merge error at index %d: %w", i, err)
				}

				output.Index(i).Set(reflect.ValueOf(newValue))
			} else {
				output.Index(i).Set(source.Index(i))
			}
		}

		return output.Interface(), nil

	default:
		return source.Interface(), nil
	}
}

func getArrayElementType(value reflect.Value) reflect.Type {
	return reflect.TypeOf(value.Interface()).Elem()
}
