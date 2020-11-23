package golden

import (
	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type ObjectTransformer struct{}

func (o *ObjectTransformer) Transform(input interface{}) (interface{}, error) {
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

func (*ObjectTransformer) setObject(obj runtime.Object) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	metaObj.SetCreationTimestamp(metav1.Time{})
	metaObj.SetUID(types.UID(""))
	metaObj.SetResourceVersion("")
	metaObj.SetGeneration(0)

	refs := metaObj.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(refs))

	for i, ref := range refs {
		newRefs[i] = ref
		newRefs[i].UID = types.UID("")
	}

	metaObj.SetOwnerReferences(newRefs)
}
