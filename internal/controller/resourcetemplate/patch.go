package resourcetemplate

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchObjectForCreate(original *unstructured.Unstructured, patch *v1beta1.WebhookPatch, meta strategicpatch.LookupPatchMeta) (*unstructured.Unstructured, error) {
	originalBuf, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal original resource: %w", err)
	}

	var modifiedBuf []byte

	switch {
	case patch.Merge != nil && patch.Merge.Raw != nil:
		modifiedBuf, err = applyMergePatch(originalBuf, patch.Merge.Raw, meta)
	case len(patch.JSONPatch) > 0:
		modifiedBuf, err = applyJSONPatch(originalBuf, patch.JSONPatch)
	default:
		modifiedBuf = originalBuf
	}

	if err != nil {
		return nil, err
	}

	var modified unstructured.Unstructured
	if err := json.Unmarshal(modifiedBuf, &modified); err != nil {
		return nil, fmt.Errorf("failed to marshal modified resource: %w", err)
	}

	modified.SetGroupVersionKind(original.GroupVersionKind())
	modified.SetName(original.GetName())
	modified.SetNamespace(original.GetNamespace())

	if modified.GetAPIVersion() == "v1" && modified.GetKind() == "Service" {
		patchService(&modified)
	}

	return &modified, nil
}

func patchService(obj *unstructured.Unstructured) {
	unstructured.RemoveNestedField(obj.Object, "spec", "clusterIP")

	ports, ok, _ := unstructured.NestedFieldNoCopy(obj.Object, "spec", "ports")
	if ok {
		if portSlice, ok := ports.([]interface{}); ok {
			for _, port := range portSlice {
				if port, ok := port.(map[string]interface{}); ok {
					delete(port, "nodePort")
				}
			}
		}
	}
}

func applyMergePatch(original, patch []byte, schema strategicpatch.LookupPatchMeta) ([]byte, error) {
	if schema == nil {
		return applyJSONMergePatch(original, patch)
	}

	return applyStrategicMergePatch(original, patch, schema)
}

func applyJSONMergePatch(original, patch []byte) ([]byte, error) {
	result, err := jsonpatch.MergePatch(original, patch)
	if err != nil {
		return nil, fmt.Errorf("failed to merge the patch: %w", err)
	}

	return result, nil
}

func applyStrategicMergePatch(original, patch []byte, schema strategicpatch.LookupPatchMeta) ([]byte, error) {
	originalMap, err := newJSONMap(original)
	if err != nil {
		return nil, err
	}

	patchMap, err := newJSONMap(patch)
	if err != nil {
		return nil, err
	}

	result, err := strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(originalMap, patchMap, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to merge the patch: %w", err)
	}

	return json.Marshal(result)
}

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

func createPatchForUpdate(original, current *unstructured.Unstructured, patch *v1beta1.WebhookPatch, schema strategicpatch.LookupPatchMeta) (client.Patch, error) {
	originalBuf, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal original resource: %w", err)
	}

	desired, err := patchObjectForCreate(original, patch, schema)
	if err != nil {
		return nil, err
	}

	desiredBuf, err := json.Marshal(desired)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal desired resource: %w", err)
	}

	currentBuf, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current resource: %w", err)
	}

	if schema == nil {
		return createJSONMergePatchForUpdate(originalBuf, desiredBuf, currentBuf)
	}

	return createStrategicMergePatchForUpdate(originalBuf, desiredBuf, currentBuf, schema)
}

func createRawPatch(patchType types.PatchType, data []byte) client.Patch {
	if bytes.Equal(data, []byte("{}")) {
		return nil
	}

	return client.RawPatch(patchType, data)
}

func createJSONMergePatchForUpdate(original, desired, current []byte) (client.Patch, error) {
	patchOriginal, err := jsonpatch.CreateMergePatch(original, desired)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	newDesired, err := applyJSONMergePatch(current, patchOriginal)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch(current, newDesired)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	return createRawPatch(types.MergePatchType, patch), nil
}

func createStrategicMergePatchForUpdate(original, desired, current []byte, schema strategicpatch.LookupPatchMeta) (client.Patch, error) {
	patchOriginal, err := strategicpatch.CreateTwoWayMergePatchUsingLookupPatchMeta(original, desired, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	newDesired, err := applyStrategicMergePatch(current, patchOriginal, schema)
	if err != nil {
		return nil, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatchUsingLookupPatchMeta(current, newDesired, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	return createRawPatch(types.StrategicMergePatchType, patch), nil
}
