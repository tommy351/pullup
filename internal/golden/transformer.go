package golden

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/testutil"
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
		o.setObject(output)

		return output, nil

	case []runtime.Object:
		return testutil.MapObjects(input, o.setObject), nil

	default:
		return input, nil
	}
}

func (ObjectTransformer) setObject(obj runtime.Object) {
	if rt, ok := obj.(*v1beta1.ResourceTemplate); ok {
		patchData := []k8s.JSONPatch{
			{Op: "remove", Path: "/webhook/metadata/creationTimestamp"},
			{Op: "remove", Path: "/webhook/metadata/generation"},
			{Op: "remove", Path: "/webhook/metadata/managedFields"},
			{Op: "remove", Path: "/webhook/metadata/namespace"},
			{Op: "remove", Path: "/webhook/metadata/resourceVersion"},
			{Op: "remove", Path: "/webhook/metadata/selfLink"},
			{Op: "remove", Path: "/webhook/metadata/uid"},
		}
		patchBuf, err := json.Marshal(patchData)
		if err != nil {
			panic(err)
		}

		patch, err := jsonpatch.DecodePatch(patchBuf)
		if err != nil {
			panic(err)
		}

		rt.Spec.Data.Raw, err = patch.Apply(rt.Spec.Data.Raw)
		if err != nil {
			panic(err)
		}
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return
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
}
