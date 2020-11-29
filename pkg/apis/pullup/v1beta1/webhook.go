package v1beta1

import extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type WebhookSpec struct {
	Patches      []WebhookPatch `json:"patches,omitempty"`
	ResourceName string         `json:"resourceName,omitempty"`
}

type WebhookStatus struct{}

type WebhookPatch struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`

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
