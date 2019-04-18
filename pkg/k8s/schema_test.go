package k8s

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = DescribeTable("GVKToTypeMeta", func(gvk schema.GroupVersionKind, expected metav1.TypeMeta) {
	Expect(GVKToTypeMeta(gvk)).To(Equal(expected))
},
	Entry("Core", schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}),
	Entry("Extensions", schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"}),
)
