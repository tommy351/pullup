package kubernetes

import (
	"github.com/ansel1/merry"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetErrorReason(err error) metav1.StatusReason {
	if e, ok := merry.Unwrap(err).(errors.APIStatus); ok {
		return e.Status().Reason
	}

	return metav1.StatusReasonUnknown
}

func IsNotFoundError(err error) bool {
	return GetErrorReason(err) == metav1.StatusReasonNotFound
}

func IsAlreadyExistError(err error) bool {
	return GetErrorReason(err) == metav1.StatusReasonAlreadyExists
}
