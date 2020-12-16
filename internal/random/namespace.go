package random

import (
	"fmt"
	"strings"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (r *NamespaceMap) SetObject(input client.Object) error {
	r.setNamespace(input, r.GetRandom(input.GetNamespace()))

	return nil
}

func (r *NamespaceMap) RestoreObject(input client.Object) error {
	r.setNamespace(input, r.GetOriginal(input.GetNamespace()))

	return nil
}
