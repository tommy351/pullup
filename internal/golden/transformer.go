package golden

import (
	"encoding/json"
	"fmt"

	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectTransformer struct{}

func (o ObjectTransformer) Transform(input interface{}) (interface{}, error) {
	switch input := input.(type) {
	case client.Object:
		output := input.DeepCopyObject().(client.Object)

		if err := o.setObject(output); err != nil {
			return nil, err
		}

		return output, nil

	case []client.Object:
		output := make([]client.Object, len(input))

		for i, obj := range input {
			output[i] = obj.DeepCopyObject().(client.Object)
		}

		return k8s.MapObjects(output, o.setObject)

	default:
		return input, nil
	}
}

func (o ObjectTransformer) setObject(obj client.Object) error {
	if rt, ok := obj.(*v1beta1.ResourceTemplate); ok {
		if err := o.setResourceTemplate(rt); err != nil {
			return err
		}
	}

	obj.SetCreationTimestamp(metav1.Time{})
	obj.SetUID("")
	obj.SetResourceVersion("")
	obj.SetGeneration(0)
	obj.SetManagedFields(nil)

	refs := obj.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(refs))

	for i, ref := range refs {
		newRefs[i] = ref
		newRefs[i].UID = ""
	}

	obj.SetOwnerReferences(newRefs)

	return nil
}

func (o ObjectTransformer) setResourceTemplate(rt *v1beta1.ResourceTemplate) error {
	if rt.Spec.Data.Raw != nil {
		data := map[string]interface{}{}

		if err := json.Unmarshal(rt.Spec.Data.Raw, &data); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}

		for _, field := range []string{
			"creationTimestamp",
			"generation",
			"managedFields",
			"namespace",
			"resourceVersion",
			"selfLink",
			"uid",
		} {
			unstructured.RemoveNestedField(data, "webhook", "metadata", field)
		}

		buf, err := json.Marshal(&data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}

		rt.Spec.Data.Raw = buf
	}

	if rt.Status.LastUpdateTime != nil {
		rt.Status.LastUpdateTime = &metav1.Time{}
	}

	return nil
}
