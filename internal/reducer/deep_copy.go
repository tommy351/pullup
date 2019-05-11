package reducer

import (
	"reflect"
)

func DeepCopy(input interface{}) interface{} {
	v := reflect.ValueOf(input)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		output := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())

		for i := 0; i < v.Len(); i++ {
			output.Index(i).Set(reflect.ValueOf(DeepCopy(v.Index(i).Interface())))
		}

		return output.Interface()

	case reflect.Map:
		output := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()

		for iter.Next() {
			output.SetMapIndex(iter.Key(), reflect.ValueOf(DeepCopy(iter.Value().Interface())))
		}

		return output.Interface()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Bool, reflect.String:
		return v.Interface()

	default:
		// TODO: Support other data types
		return nil
	}
}
