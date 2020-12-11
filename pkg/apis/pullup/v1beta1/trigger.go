package v1beta1

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=all;pullup

type Trigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TriggerStatus `json:"status,omitempty"`
	Spec   TriggerSpec   `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type TriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Trigger `json:"items"`
}

type TriggerSpec struct {
	ResourceName string         `json:"resourceName"`
	Patches      []TriggerPatch `json:"patches,omitempty"`
	Schema       *extv1.JSON    `json:"schema,omitempty"`
}

type TriggerStatus struct{}

type TriggerPatch struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	SourceName string `json:"sourceName,omitempty"`
	TargetName string `json:"targetName,omitempty"`

	// +kubebuilder:validation:Type=object
	Merge *extv1.JSON `json:"merge,omitempty"`

	JSONPatch []JSONPatch `json:"jsonPatch,omitempty"`
}

type JSONPatch struct {
	Operation JSONPatchOperation `json:"op"`
	Path      string             `json:"path"`
	From      string             `json:"from,omitempty"`
	Value     *extv1.JSON        `json:"value,omitempty"`
}

// +kubebuilder:validation:Enum=add;remove;replace;copy;move;test
type JSONPatchOperation string

const (
	JSONPatchOpAdd     JSONPatchOperation = "add"
	JSONPatchOpRemove  JSONPatchOperation = "remove"
	JSONPatchOpReplace JSONPatchOperation = "replace"
	JSONPatchOpCopy    JSONPatchOperation = "copy"
	JSONPatchOpMove    JSONPatchOperation = "move"
	JSONPatchOpTest    JSONPatchOperation = "test"
)
