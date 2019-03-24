package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
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
	Resources []json.RawMessage `json:"resources"`
	GitHub    *GitHubOptions    `json:"github"`
}

type WebhookStatus struct{}

type GitHubOptions struct {
	Secret string `json:"secret"`
}
