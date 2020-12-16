package v1beta1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionApply  = "apply"
	ActionDelete = "delete"
)

func IsActionValid(action string) bool {
	switch action {
	case ActionCreate, ActionUpdate, ActionApply, ActionDelete:
		return true
	}

	return false
}

type SecretValue struct {
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type ObjectReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name"`
}

func (in ObjectReference) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(in.APIVersion, in.Kind)
}

func (in ObjectReference) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: in.Namespace,
		Name:      in.Name,
	}
}

func (in ObjectReference) String() string {
	return fmt.Sprintf("%s/%s %s/%s", in.APIVersion, in.Kind, in.Namespace, in.Name)
}
