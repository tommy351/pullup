package random

import (
	"fmt"
	"strings"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

func Namespace() string {
	return fmt.Sprintf("pullup-test-%s", rand.String(5))
}

type NamespaceMap struct {
	m map[string]string
}

func NewNamespaceMap() *NamespaceMap {
	return &NamespaceMap{
		m: map[string]string{},
	}
}

func (r *NamespaceMap) GetRandom(original string) string {
	result, ok := r.m[original]

	if !ok {
		result = Namespace()
		r.m[original] = result
	}

	return result
}

func (r *NamespaceMap) GetOriginal(random string) string {
	for k, v := range r.m {
		if v == random {
			return k
		}
	}

	return ""
}

func (r *NamespaceMap) setNamespace(obj corev1.Object, newNamespace string) {
	oldNamespace := obj.GetNamespace()
	obj.SetNamespace(newNamespace)
	obj.SetSelfLink(strings.ReplaceAll(obj.GetSelfLink(), oldNamespace, newNamespace))

	if rt, ok := obj.(*v1beta1.ResourceTemplate); ok {
		if ref := rt.Spec.TriggerRef; ref != nil {
			ref.Namespace = newNamespace
		}

		for i := range rt.Status.Active {
			rt.Status.Active[i].Namespace = newNamespace
		}
	}
}

func (r *NamespaceMap) SetObject(input runtime.Object) error {
	obj, err := meta.Accessor(input)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}

	r.setNamespace(obj, r.GetRandom(obj.GetNamespace()))

	return nil
}

func (r *NamespaceMap) RestoreObject(input runtime.Object) error {
	obj, err := meta.Accessor(input)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}

	r.setNamespace(obj, r.GetOriginal(obj.GetNamespace()))

	return nil
}
