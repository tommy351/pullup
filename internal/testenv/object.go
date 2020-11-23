package testenv

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
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
				return fmt.Errorf("failed to parse the API version: %w", err)
			}

			gvk := gv.WithKind(ref.Kind)
			refObj, err := scheme.New(gvk)
			if err != nil {
				return fmt.Errorf("failed to create the scheme: %w", err)
			}

			key := types.NamespacedName{
				Namespace: metaObj.GetNamespace(),
				Name:      ref.Name,
			}

			if err := client.Get(ctx, key, refObj); err != nil {
				return fmt.Errorf("failed to get the resource: %w", err)
			}

			refMeta, err := meta.Accessor(refObj)
			if err != nil {
				return fmt.Errorf("failed to get the accessor: %w", err)
			}

			newRefs[i].UID = refMeta.GetUID()
		}
	}

	metaObj.SetOwnerReferences(newRefs)

	return nil
}

func waitForObject(ctx context.Context, obj runtime.Object) error {
	client := GetClient()
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("failed to get the accessor: %w", err)
	}

	key := types.NamespacedName{
		Namespace: metaObj.GetNamespace(),
		Name:      metaObj.GetName(),
	}

	return retry.OnError(retry.DefaultRetry, errors.IsNotFound, func() error {
		return client.Get(ctx, key, obj.DeepCopyObject())
	})
}

func createAndWaitObject(ctx context.Context, obj runtime.Object) error {
	if err := GetClient().Create(ctx, obj); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}

		return fmt.Errorf("failed to create the object: %w", err)
	}

	return waitForObject(ctx, obj)
}

func createNamespace(ctx context.Context, namespace string) error {
	objects := []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: namespace,
			},
			AutomountServiceAccountToken: pointer.BoolPtr(false),
		},
	}

	for _, obj := range objects {
		if err := createAndWaitObject(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func CreateObjects(objects []runtime.Object) error {
	ctx := context.Background()

	for _, obj := range objects {
		metaObj, err := meta.Accessor(obj)
		if err != nil {
			return fmt.Errorf("failed to get the accessor: %w", err)
		}

		if ns := metaObj.GetNamespace(); ns != "" {
			if err := createNamespace(ctx, ns); err != nil {
				return err
			}
		}

		if err := setOwnerReferences(ctx, obj); err != nil {
			return err
		}

		if err := createAndWaitObject(ctx, obj); err != nil {
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
				return fmt.Errorf("failed to delete the resource: %w", err)
			}
		}
	}

	return nil
}
