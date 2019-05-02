package resourceset

import "github.com/google/go-cmp/cmp"

func equal(a, b interface{}) bool {
	return cmp.Equal(a, b, cmp.Transformer("RemoveNilInMap", func(input map[string]interface{}) map[string]interface{} {
		output := make(map[string]interface{})

		for k, v := range input {
			if v != nil {
				output[k] = v
			}
		}

		return output
	}))
}
