package v1beta1

import extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type EventSourceSpec struct {
	Action   string               `json:"action,omitempty"`
	Triggers []EventSourceTrigger `json:"triggers,omitempty"`
}

type EventSourceStatus struct{}

type EventSourceTrigger struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Transform *extv1.JSON `json:"transform,omitempty"`
}

type EventSourceFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}
