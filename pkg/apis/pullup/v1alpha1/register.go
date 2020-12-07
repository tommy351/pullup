package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// nolint: gochecknoglobals
var (
	GroupVersion  = schema.GroupVersion{Group: "pullup.dev", Version: "v1alpha1"}
	SchemeBuilder = runtime.NewSchemeBuilder(AddKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&Webhook{},
		&WebhookList{},
		&ResourceSet{},
		&ResourceSetList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)

	return nil
}
