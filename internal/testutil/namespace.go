package testutil

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

type Map struct {
	keyMap   map[string]string
	valueMap map[string]string
}

func NewMap() *Map {
	return &Map{
		keyMap:   map[string]string{},
		valueMap: map[string]string{},
	}
}

func (m *Map) Set(key, value string) (string, bool) {
	if v, ok := m.valueMap[key]; ok {
		return v, false
	}

	m.keyMap[value] = key
	m.valueMap[key] = value
	return value, true
}

func (m *Map) Value(key string) string {
	return m.valueMap[key]
}

func (m *Map) Key(value string) string {
	return m.keyMap[value]
}

func replaceNamespace(obj metav1.Object, oldNamespace, newNamespace string) {
	obj.SetNamespace(newNamespace)
	obj.SetSelfLink(strings.Replace(obj.GetSelfLink(), "/namespaces/"+oldNamespace+"/", "/namespaces/"+newNamespace+"/", -1))
}

func SetRandomNamespace(m *Map, input []runtime.Object) []runtime.Object {
	return MapObjects(input, func(object runtime.Object) {
		o, err := meta.Accessor(object)

		if err == nil {
			ns, _ := m.Set(o.GetNamespace(), rand.String(8))
			replaceNamespace(o, o.GetNamespace(), ns)
		}
	})
}

func RestoreNamespace(m *Map, input []runtime.Object) []runtime.Object {
	return MapObjects(input, func(object runtime.Object) {
		o, err := meta.Accessor(object)

		if err == nil {
			replaceNamespace(o, o.GetNamespace(), m.Key(o.GetNamespace()))
		}
	})
}
