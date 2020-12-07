package resourcetemplate

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func applyJSONPatch(original []byte, patch []v1beta1.JSONPatch) ([]byte, error) {
	patchBuf, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON patch: %w", err)
	}

	jp, err := jsonpatch.DecodePatch(patchBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON patch: %w", err)
	}

	return jp.Apply(original)
}

func newRawPatch(patchType types.PatchType, data []byte) client.Patch {
	if bytes.Equal(data, []byte("{}")) {
		return nil
	}

	return client.RawPatch(patchType, data)
}

func newJSONMergePatchForUpdate(original, desired, current []byte) (client.Patch, error) {
	patchOriginal, err := jsonpatch.CreateMergePatch(original, desired)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	newDesired, err := jsonpatch.MergePatch(current, patchOriginal)
	if err != nil {
		return nil, fmt.Errorf("failed to merge patch: %w", err)
	}

	patch, err := jsonpatch.CreateMergePatch(current, newDesired)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	return newRawPatch(types.MergePatchType, patch), nil
}

func newStrategicMergePatchForUpdate(original, desired, current []byte, dataStruct interface{}) (client.Patch, error) {
	patchOriginal, err := strategicpatch.CreateTwoWayMergePatch(original, desired, dataStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	newDesired, err := strategicpatch.StrategicMergePatch(current, patchOriginal, dataStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to merge patch: %w", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(current, newDesired, dataStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	return newRawPatch(types.StrategicMergePatchType, patch), nil
}
