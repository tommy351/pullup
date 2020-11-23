package merge

import (
	"reflect"
)

type Func func(input, source interface{}) (interface{}, error)

// nolint: gochecknoglobals
var (
	// Any is a default merge func for any types.
	Any = makeAny()
)

func makeAny() Func {
	merger := new(Merger)
	merger.Func = merger.Merge

	return merger.Func
}

type Merger struct {
	Func Func
}

func (m *Merger) Merge(input, source interface{}) (interface{}, error) {
	sv := reflect.ValueOf(source)

	if !sv.IsValid() {
		return nil, nil
	}

	iv := reflect.ValueOf(input)

	switch iv.Kind() {
	case reflect.Array, reflect.Slice:
		return m.mergeArray(iv, sv)
	case reflect.Map:
		return m.mergeMap(iv, sv)
	default:
		return sv.Interface(), nil
	}
}

func getUnionType(a, b reflect.Type) reflect.Type {
	if a == b {
		return a
	}

	return reflect.TypeOf(new(interface{})).Elem()
}
