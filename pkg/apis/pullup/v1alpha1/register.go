package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// nolint: gochecknoglobals
var (
	SchemeGroupVersion = schema.GroupVersion{Group: "pullup.dev", Version: "v1alpha1"}
	SchemeBuilder      = &scheme.Builder{GroupVersion: SchemeGroupVersion}
	AddToScheme        = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
