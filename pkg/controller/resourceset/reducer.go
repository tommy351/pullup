package resourceset

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/tommy351/pullup/pkg/reducer"
	"github.com/tommy351/pullup/pkg/reducer/collection"
	"github.com/tommy351/pullup/pkg/reducer/merge"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
)

func deleteKeys(keys []string) reducer.Interface {
	dict := map[interface{}]struct{}{}

	for _, k := range keys {
		dict[k] = struct{}{}
	}

	filter := reducer.FilterKey(func(key interface{}) (bool, error) {
		_, ok := dict[key]
		return !ok, nil
	})

	return reducer.Func(func(input interface{}) (interface{}, error) {
		output, err := filter.Reduce(input)

		if err != nil {
			if !xerrors.Is(err, reducer.ErrNotMap) {
				return nil, err
			}

			return input, nil
		}

		return output, nil
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

func newTemplateReducer(data interface{}) reducer.Interface {
	return reducer.DeepMapValue(func(value interface{}) (interface{}, error) {
		s, ok := value.(string)

		if !ok {
			return value, nil
		}

		tmpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(s)

		if err != nil {
			return nil, xerrors.Errorf("failed to parse template: %w", err)
		}

		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, xerrors.Errorf("failed to execute template: %w", err)
		}

		return buf.String(), nil
	})
}

func normalizeValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case types.UID:
		return string(v), nil
	default:
		return v, nil
	}
}