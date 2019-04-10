package k8s

import (
	"strings"

	"github.com/jinzhu/inflection"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func KindToResource(kind string) string {
	return inflection.Plural(strings.ToLower(kind))
}

func ParseGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)

	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: KindToResource(kind),
	}, nil
}

func GVKToTypeMeta(gvk schema.GroupVersionKind) metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
	}
}
