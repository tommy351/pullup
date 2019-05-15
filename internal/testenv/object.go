package testenv

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func setOwnerReferences(ctx context.Context, obj runtime.Object) error {
	client := GetClient()
	scheme := GetScheme()
	metaObj, err := meta.Accessor(obj)

	if err != nil {
		return nil
	}

	refs := metaObj.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(refs))

	for i, ref := range refs {
		newRefs[i] = ref

		if ref.UID == "" {
			gv, err := schema.ParseGroupVersion(ref.APIVersion)

			if err != nil {
				return err
			}

			gvk := gv.WithKind(ref.Kind)
			refObj, err := scheme.New(gvk)

			if err != nil {
				return err
			}

			key := types.NamespacedName{
				Namespace: metaObj.GetNamespace(),
				Name:      ref.Name,
			}

			if err := client.Get(ctx, key, refObj); err != nil {
				return err
			}

			refMeta, err := meta.Accessor(refObj)

			if err != nil {
				return err
			}

			newRefs[i].UID = refMeta.GetUID()
		}
	}

	metaObj.SetOwnerReferences(newRefs)
	return nil
}

func CreateObjects(objects []runtime.Object) error {
	ctx := context.Background()
	client := GetClient()

	for _, obj := range objects {
		if err := setOwnerReferences(ctx, obj); err != nil {
			return err
		}

		if err := client.Create(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func DeleteObjects(objects []runtime.Object) error {
	ctx := context.Background()
	client := GetClient()

	for _, obj := range objects {
		if err := client.Delete(ctx, obj); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}
