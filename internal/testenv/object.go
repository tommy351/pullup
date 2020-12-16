package testenv

import (
	"context"
	"fmt"

	"github.com/tommy351/pullup/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setOwnerReferences(ctx context.Context, obj client.Object) error {
	client := GetClient()
	scheme := GetScheme()
	refs := obj.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(refs))

	for i, ref := range refs {
		newRefs[i] = ref

		if ref.UID == "" {
			gv, err := schema.ParseGroupVersion(ref.APIVersion)
			if err != nil {
				return fmt.Errorf("failed to parse the API version: %w", err)
			}

			gvk := gv.WithKind(ref.Kind)
			refObj, err := k8s.NewEmptyObject(scheme, gvk)
			if err != nil {
				return fmt.Errorf("failed to create the scheme: %w", err)
			}

			key := types.NamespacedName{
				Namespace: obj.GetNamespace(),
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

	obj.SetOwnerReferences(newRefs)

	return nil
}

func waitForObject(ctx context.Context, obj client.Object) error {
	c := GetClient()
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	return retry.OnError(retry.DefaultRetry, errors.IsNotFound, func() error {
		return c.Get(ctx, key, obj.DeepCopyObject().(client.Object))
	})
}

func createAndWaitObject(ctx context.Context, obj client.Object) error {
	if err := GetClient().Create(ctx, obj); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}

		return fmt.Errorf("failed to create the object: %w", err)
	}

	return waitForObject(ctx, obj)
}

func createNamespace(ctx context.Context, namespace string) error {
	objects := []client.Object{
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

func CreateObjects(objects []client.Object) error {
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

func DeleteObjects(objects []client.Object) error {
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
