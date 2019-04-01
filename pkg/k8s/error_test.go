package k8s

import (
	"errors"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// nolint: gochecknoglobals
var notFoundErr = apierrors.NewNotFound(schema.GroupResource{}, "")

var _ = DescribeTable("GetErrorReason", func(input error, expected metav1.StatusReason) {
	Expect(GetErrorReason(input)).To(Equal(expected))
},
	Entry("Not APIStatus", errors.New(""), metav1.StatusReasonUnknown),
	Entry("Naked StatusError", notFoundErr, metav1.StatusReasonNotFound),
	Entry("Wrapped StatusError", xerrors.Errorf("k8s error: %w", notFoundErr), metav1.StatusReasonNotFound),
)

var _ = DescribeTable("IsNotFoundError", func(input error, expected bool) {
	Expect(IsNotFoundError(input)).To(Equal(expected))
},
	Entry("true", apierrors.NewNotFound(schema.GroupResource{}, ""), true),
	Entry("false", errors.New(""), false),
)

var _ = DescribeTable("IsAlreadyExistError", func(input error, expected bool) {
	Expect(IsAlreadyExistError(input)).To(Equal(expected))
},
	Entry("true", apierrors.NewAlreadyExists(schema.GroupResource{}, ""), true),
	Entry("false", errors.New(""), false),
)
