package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "pullup.dev"

// nolint: gochecknoglobals
var (
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}
)

func Kind(kind string) schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(kind)
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Webhook{},
		&WebhookList{},
		&ResourceSet{},
		&ResourceSetList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
