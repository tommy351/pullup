package v1beta1

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type HTTPWebhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status HTTPWebhookStatus `json:"status,omitempty"`
	Spec   HTTPWebhookSpec   `json:"spec,omitempty"`
}

func (in *HTTPWebhook) GetSpec() WebhookSpec {
	return in.Spec.WebhookSpec
}

// +kubebuilder:object:root=true

type HTTPWebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HTTPWebhook `json:"items"`
}

type HTTPWebhookSpec struct {
	WebhookSpec `json:",inline"`

	Schema      *extv1.JSON  `json:"schema,omitempty"`
	SecretToken *SecretValue `json:"secretToken,omitempty"`
}

type HTTPWebhookStatus struct {
	WebhookStatus `json:",inline"`
}
