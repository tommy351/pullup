package random

import (
	"fmt"
	"strings"

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
	obj.SetSelfLink(strings.Replace(obj.GetSelfLink(), oldNamespace, newNamespace, -1))
}

func (r *NamespaceMap) SetObject(input runtime.Object) {
	obj, err := meta.Accessor(input)

	if err == nil {
		r.setNamespace(obj, r.GetRandom(obj.GetNamespace()))
	}
}

func (r *NamespaceMap) RestoreObject(input runtime.Object) {
	obj, err := meta.Accessor(input)

	if err == nil {
		r.setNamespace(obj, r.GetOriginal(obj.GetNamespace()))
	}
}
