package reducer

import "github.com/tommy351/pullup/internal/reducer/merge"

func Merge(source interface{}) Interface {
	return MergeWith(source, merge.Any)
}

func MergeWith(source interface{}, fn merge.Func) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		return fn(input, source)
	})
}
