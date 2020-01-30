package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Status WebhookStatus `json:"status"`
	Spec   WebhookSpec   `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Webhook `json:"items"`
}

type WebhookSpec struct {
	Resources    []json.RawMessage   `json:"resources,omitempty"`
	Repositories []WebhookRepository `json:"repositories"`
	ResourceName string              `json:"resourceName"`
}

type WebhookRepository struct {
	Type     string        `json:"type"`
	Name     string        `json:"name"`
	Branches WebhookFilter `json:"branches"`
}

type WebhookFilter struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

type WebhookStatus struct{}

// nolint: gochecknoinits
func init() {
	SchemeBuilder.Register(&Webhook{}, &WebhookList{})
}
