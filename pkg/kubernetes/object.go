package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

type Object struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func DecodeObject(data []byte) (*Object, error) {
	var obj Object

	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}
