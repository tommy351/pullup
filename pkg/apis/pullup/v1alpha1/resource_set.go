package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ResourceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Status ResourceSetStatus `json:"status"`
	Spec   ResourceSetSpec   `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ResourceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ResourceSet `json:"items"`
}

type ResourceSetSpec struct {
	Resources []json.RawMessage `json:"resources"`
	Number    int               `json:"number"`
	Base      Commit            `json:"base"`
	Head      Commit            `json:"head"`
}

type ResourceSetStatus struct {
}

type Commit struct {
	Ref string `json:"ref,omitempty"`
	SHA string `json:"sha,omitempty"`
}
