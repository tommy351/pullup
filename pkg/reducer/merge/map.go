package merge

import (
	"reflect"

	"golang.org/x/xerrors"
)

func (m *Merger) mergeMap(input, source reflect.Value) (interface{}, error) {
	if source.Kind() != reflect.Map {
		return source.Interface(), nil
	}

	output := reflect.MakeMap(reflect.MapOf(
		getUnionType(input.Type().Key(), source.Type().Key()),
		getUnionType(input.Type().Elem(), source.Type().Elem()),
	))

	// Set input values
	iter := input.MapRange()

	for iter.Next() {
		output.SetMapIndex(iter.Key(), iter.Value())
	}

	// Merge source values
	iter = source.MapRange()

	for iter.Next() {
		if inputValue := output.MapIndex(iter.Key()); inputValue.IsValid() {
			newValue, err := m.Func(inputValue.Interface(), iter.Value().Interface())

			if err != nil {
				return nil, xerrors.Errorf("merge error at key %v: %w", iter.Key().Interface(), err)
			}

			output.SetMapIndex(iter.Key(), reflect.ValueOf(newValue))
		} else {
			output.SetMapIndex(iter.Key(), iter.Value())
		}
	}

	return output.Interface(), nil
}
