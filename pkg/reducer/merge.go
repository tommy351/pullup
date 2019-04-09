package reducer

import (
	"golang.org/x/xerrors"
	"k8s.io/utils/integer"
)

type Merge struct {
	Source interface{}
}

func (m Merge) Reduce(input interface{}) (interface{}, error) {
	return mergeValue(input, m.Source)
}

func mergeArray(base, patch []interface{}) ([]interface{}, error) {
	if isNamedArray(base) && isNamedArray(patch) {
		return mergeNamedArray(base, patch)
	}

	output := make([]interface{}, integer.IntMax(len(base), len(patch)))

	copy(output, base)

	for i, v := range patch {
		newValue, err := mergeValue(getArrayElement(base, i), v)

		if err != nil {
			return nil, xerrors.Errorf("failed to merge array at index %d: %w", i, err)
		}

		output[i] = newValue
	}

	return output, nil
}

func mergeNamedArray(base, patch []interface{}) ([]interface{}, error) {
	nameMap := map[string]interface{}{}

	for _, v := range base {
		name, _ := getNameFromMap(v)
		nameMap[name] = v
	}

	for _, v := range patch {
		name, _ := getNameFromMap(v)
		newValue, err := mergeValue(nameMap[name], v)

		if err != nil {
			return nil, xerrors.Errorf("failed to merge array with name %s: %w", name, err)
		}

		nameMap[name] = newValue
	}

	output := make([]interface{}, 0, len(nameMap))

	for _, v := range nameMap {
		output = append(output, v)
	}

	return output, nil
}

func isNamedArray(input []interface{}) bool {
	for _, x := range input {
		if _, ok := getNameFromMap(x); !ok {
			return false
		}
	}

	return true
}

func getNameFromMap(input interface{}) (string, bool) {
	if m, ok := input.(map[string]interface{}); ok {
		if name, ok := m["name"]; ok {
			if s, ok := name.(string); ok {
				return s, true
			}
		}
	}

	return "", false
}

func getArrayElement(arr []interface{}, i int) interface{} {
	if i < len(arr) {
		return arr[i]
	}

	return nil
}

func mergeMap(base, patch map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{}, integer.IntMax(len(base), len(patch)))

	for k, v := range base {
		output[k] = v
	}

	for k, v := range patch {
		v, err := mergeValue(base[k], v)

		if err != nil {
			return nil, xerrors.Errorf("failed to merge key %s: %w", k, err)
		}

		output[k] = v
	}

	return output, nil
}

func mergeValue(base, patch interface{}) (interface{}, error) {
	switch patch := patch.(type) {
	case Interface:
		if base == nil {
			return nil, nil
		}

		return patch.Reduce(base)

	case []interface{}:
		baseArr, ok := base.([]interface{})

		if !ok {
			baseArr = []interface{}{}
		}

		return mergeArray(baseArr, patch)

	case map[string]interface{}:
		baseMap, ok := base.(map[string]interface{})

		if !ok {
			baseMap = map[string]interface{}{}
		}

		return mergeMap(baseMap, patch)

	default:
		return patch, nil
	}
}
