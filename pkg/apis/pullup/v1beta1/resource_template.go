package v1beta1

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
	Patches []WebhookPatch `json:"patches,omitempty"`
	Data    extv1.JSON     `json:"data,omitempty"`
}

type ResourceTemplateStatus struct {
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	ResourceStatuses []ResourceStatus `json:"resourceStatuses,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type ResourceStatus struct {
	Name            string      `json:"name"`
	LastAppliedTime metav1.Time `json:"lastAppliedTime"`
}
