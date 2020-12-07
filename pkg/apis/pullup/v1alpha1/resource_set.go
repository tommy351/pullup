package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type ResourceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status ResourceSetStatus `json:"status,omitempty"`
	Spec   ResourceSetSpec   `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type ResourceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceSet `json:"items"`
}

type ResourceSetSpec struct {
	Resources []WebhookResource `json:"resources"`
	// +optional
	Number int `json:"number"`
	// +optional
	Base Commit `json:"base"`
	// +optional
	Head Commit `json:"head"`
}

type ResourceSetStatus struct {
}

type Commit struct {
	// +optional
	Ref string `json:"ref,omitempty"`
	// +optional
	SHA string `json:"sha,omitempty"`
}
