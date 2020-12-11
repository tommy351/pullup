package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +kubebuilder:validation:Enum=create;update;apply;delete
type WebhookAction string

const (
	WebhookActionCreate WebhookAction = "create"
	WebhookActionUpdate WebhookAction = "update"
	WebhookActionApply  WebhookAction = "apply"
	WebhookActionDelete WebhookAction = "delete"
)

type WebhookSpec struct {
	Patches      []WebhookPatch `json:"patches,omitempty"`
	ResourceName string         `json:"resourceName,omitempty"`
}

type WebhookStatus struct{}

type WebhookPatch struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	SourceName string `json:"sourceName,omitempty"`
	TargetName string `json:"targetName,omitempty"`

	// +kubebuilder:validation:Type=object
	Merge *extv1.JSON `json:"merge,omitempty"`

	JSONPatch []JSONPatch `json:"jsonPatch,omitempty"`
}

type WebhookFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

type JSONPatch struct {
	// +kubebuilder:validation:Enum=add;remove;replace;copy;move;test
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	From      string      `json:"from,omitempty"`
	Value     *extv1.JSON `json:"value,omitempty"`
}

type SecretValue struct {
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type ObjectReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

func (in ObjectReference) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(in.APIVersion, in.Kind)
}
