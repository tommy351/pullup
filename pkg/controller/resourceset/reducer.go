package resourceset

import (
	"github.com/tommy351/pullup/pkg/reducer"
	"github.com/tommy351/pullup/pkg/reducer/collection"
	"github.com/tommy351/pullup/pkg/reducer/merge"
)

func deleteKeys(keys []string) reducer.Interface {
	dict := map[interface{}]struct{}{}

	for _, k := range keys {
		dict[k] = struct{}{}
	}

	return reducer.FilterKey(func(key interface{}) (bool, error) {
		_, ok := dict[key]
		return !ok, nil
	})
}

func mergeResource(data interface{}) reducer.Interface {
	merger := new(merge.Merger)
	merger.Func = func(input, source interface{}) (interface{}, error) {
		inputArr, ok := collection.NewNamedArray(input)

		if !ok {
			return merger.Merge(input, source)
		}

		srcArr, ok := collection.NewNamedArray(source)

		if !ok {
			return merger.Merge(input, source)
		}

		iter := srcArr.Iterate()

		for iter.Next() {
			inputValue, ok := inputArr.Get(iter.Key())

			if !ok {
				inputArr.Set(iter.Key(), iter.Value())
				continue
			}

			newValue, err := merger.Merge(inputValue, iter.Value())

			if err != nil {
				return nil, err
			}

			inputArr.Set(iter.Key(), newValue)
		}

		return inputArr.ToArray(), nil
	}

	return reducer.MergeWith(data, merger.Func)
}
