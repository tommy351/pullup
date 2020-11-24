package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status WebhookStatus `json:"status,omitempty"`
	Spec   WebhookSpec   `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Webhook `json:"items"`
}

type WebhookSpec struct {
	Resources    []WebhookResource   `json:"resources,omitempty"`
	Repositories []WebhookRepository `json:"repositories,omitempty"`
	ResourceName string              `json:"resourceName,omitempty"`
}

type WebhookRepository struct {
	// +kubebuilder:validation:Enum=github
	// +optional
	Type     string        `json:"type"`
	Name     string        `json:"name"`
	Branches WebhookFilter `json:"branches,omitempty"`
}

type WebhookFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

type WebhookStatus struct{}

// +kubebuilder:validation:XEmbeddedResource
// +kubebuilder:validation:XPreserveUnknownFields

type WebhookResource struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`

	unstructured.Unstructured `json:"-"`
}

func (in WebhookResource) MarshalJSON() ([]byte, error) {
	return in.Unstructured.MarshalJSON()
}

func (in *WebhookResource) UnmarshalJSON(data []byte) error {
	if err := in.Unstructured.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("failed to unmarshal WebhookResource: %w", err)
	}

	in.APIVersion = in.Unstructured.GetAPIVersion()
	in.Kind = in.Unstructured.GetKind()
	in.Metadata = meta.AsPartialObjectMetadata(&in.Unstructured).ObjectMeta

	return nil
}

// nolint: gochecknoinits
func init() {
	SchemeBuilder.Register(&Webhook{}, &WebhookList{})
}
