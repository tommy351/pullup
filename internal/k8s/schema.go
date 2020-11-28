package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LabelWebhookName       = "webhook-name"
	LabelPullRequestNumber = "pull-request-number"
)

type JSONPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func GVKToTypeMeta(gvk schema.GroupVersionKind) metav1.TypeMeta {
	var meta metav1.TypeMeta
	meta.APIVersion, meta.Kind = gvk.ToAPIVersionAndKind()

	return meta
}
