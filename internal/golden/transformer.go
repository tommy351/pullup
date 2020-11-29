package golden

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ObjectTransformer struct{}

func (o ObjectTransformer) Transform(input interface{}) (interface{}, error) {
	switch input := input.(type) {
	case runtime.Object:
		output := input.DeepCopyObject()

		if err := o.setObject(output); err != nil {
			return nil, err
		}

		return output, nil

	case []runtime.Object:
		return k8s.MapObjects(input, o.setObject)

	default:
		return input, nil
	}
}

func (ObjectTransformer) setObject(obj runtime.Object) error {
	if rt, ok := obj.(*v1beta1.ResourceTemplate); ok {
		patchData := []v1beta1.JSONPatch{
			{Operation: "remove", Path: "/webhook/metadata/creationTimestamp"},
			{Operation: "remove", Path: "/webhook/metadata/generation"},
			{Operation: "remove", Path: "/webhook/metadata/managedFields"},
			{Operation: "remove", Path: "/webhook/metadata/namespace"},
			{Operation: "remove", Path: "/webhook/metadata/resourceVersion"},
			{Operation: "remove", Path: "/webhook/metadata/selfLink"},
			{Operation: "remove", Path: "/webhook/metadata/uid"},
		}
		patchBuf, err := json.Marshal(patchData)
		if err != nil {
			return fmt.Errorf("failed to marshal patch: %w", err)
		}

		patch, err := jsonpatch.DecodePatch(patchBuf)
		if err != nil {
			return fmt.Errorf("failed to decode patch: %w", err)
		}

		rt.Spec.Data.Raw, err = patch.Apply(rt.Spec.Data.Raw)
		if err != nil {
			return fmt.Errorf("failed to apply patch: %w", err)
		}
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return nil
	}

	metaObj.SetCreationTimestamp(metav1.Time{})
	metaObj.SetUID("")
	metaObj.SetResourceVersion("")
	metaObj.SetGeneration(0)
	metaObj.SetManagedFields(nil)

	refs := metaObj.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(refs))

	for i, ref := range refs {
		newRefs[i] = ref
		newRefs[i].UID = ""
	}

	metaObj.SetOwnerReferences(newRefs)

	return nil
}
