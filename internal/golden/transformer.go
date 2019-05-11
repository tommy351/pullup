package golden

import (
	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Transformer interface {
	Transform(interface{}) interface{}
}

type ObjectTransformer struct {
}

func (o *ObjectTransformer) Transform(input interface{}) interface{} {
	switch input := input.(type) {
	case runtime.Object:
		output := input.DeepCopyObject()
		o.setObject(output)
		return output

	case []runtime.Object:
		return testutil.MapObjects(input, o.setObject)

	default:
		return input
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

	refs := metaObj.GetOwnerReferences()

	for i := range refs {
		refs[i].UID = types.UID("")
	}
}
