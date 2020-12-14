package jsonutil

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func AddMapKey(input []byte, key string, value interface{}) ([]byte, error) {
	return AddMapKeys(input, map[string]interface{}{key: value})
}

func AddMapKeys(input []byte, values map[string]interface{}) ([]byte, error) {
	patches := make([]v1beta1.JSONPatch, 0, len(values))

	for k, v := range values {
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}

		patches = append(patches, v1beta1.JSONPatch{
			Operation: v1beta1.JSONPatchOpAdd,
			Path:      "/" + k,
			Value:     &extv1.JSON{Raw: buf},
		})
	}

	patchBuf, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	patch, err := jsonpatch.DecodePatch(patchBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json patch: %w", err)
	}

	result, err := patch.Apply(input)
	if err != nil {
		return nil, fmt.Errorf("failed to apply json patch: %w", err)
	}

	return result, nil
}

func PickKeys(input []byte, keys []string) ([]byte, error) {
	var inputMap map[string]json.RawMessage

	if err := json.Unmarshal(input, &inputMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	outputMap := map[string]json.RawMessage{}

	for _, key := range keys {
		if v, ok := inputMap[key]; ok {
			outputMap[key] = v
		}
	}

	return json.Marshal(outputMap)
}
