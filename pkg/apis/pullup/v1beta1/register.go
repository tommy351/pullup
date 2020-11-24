package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// nolint: gochecknoglobals
var (
	GroupVersion  = schema.GroupVersion{Group: "pullup.dev", Version: "v1beta1"}
	SchemeBuilder = runtime.NewSchemeBuilder(AddKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func AddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&ResourceTemplate{},
		&ResourceTemplateList{},
		&GitHubWebhook{},
		&GitHubWebhookList{},
		&HTTPWebhook{},
		&HTTPWebhookList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)

	return nil
}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}
