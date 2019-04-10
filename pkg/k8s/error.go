package k8s

import (
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetErrorReason(err error) metav1.StatusReason {
	if err == nil {
		return metav1.StatusReasonUnknown
	}

	var status errors.APIStatus

	if xerrors.As(err, &status) {
		return status.Status().Reason
	}

	return metav1.StatusReasonUnknown
}

func IsNotFoundError(err error) bool {
	return GetErrorReason(err) == metav1.StatusReasonNotFound
}

func IsAlreadyExistError(err error) bool {
	return GetErrorReason(err) == metav1.StatusReasonAlreadyExists
}
