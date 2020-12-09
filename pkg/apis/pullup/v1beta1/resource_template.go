package v1beta1

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DataKeyEvent    = "event"
	DataKeyWebhook  = "webhook"
	DataKeyResource = "resource"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=all;pullup
// +kubebuilder:printcolumn:name="Webhook Kind",type=string,JSONPath=`.spec.webhookRef.kind`
// +kubebuilder:printcolumn:name="Webhook Name",type=string,JSONPath=`.spec.webhookRef.name`
// +kubebuilder:printcolumn:name="Last Update",type=date,JSONPath=`.status.lastUpdateTime`

type ResourceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status ResourceTemplateStatus `json:"status,omitempty"`
	Spec   ResourceTemplateSpec   `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type ResourceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceTemplate `json:"items"`
}

type ResourceTemplateSpec struct {
	WebhookRef *ObjectReference `json:"webhookRef,omitempty"`
	Patches    []WebhookPatch   `json:"patches,omitempty"`
	Data       extv1.JSON       `json:"data,omitempty"`
}

type ResourceTemplateStatus struct {
	LastUpdateTime *metav1.Time        `json:"lastUpdateTime,omitempty"`
	Active         []ResourceReference `json:"active,omitempty"`
}

type ResourceReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}
