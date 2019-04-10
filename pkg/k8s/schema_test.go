package k8s

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = DescribeTable("KindToResource", func(input, expected string) {
	Expect(KindToResource(input)).To(Equal(expected))
},
	Entry("Capitalized singular", "Pod", "pods"),
	Entry("Plural", "pods", "pods"),
)

var _ = DescribeTable("ParseGVR", func(apiVersion, kind string, assert func(schema.GroupVersionResource, error)) {
	assert(ParseGVR(apiVersion, kind))
},
	Entry("Core", "v1", "Pod", func(gvr schema.GroupVersionResource, err error) {
		Expect(gvr).To(Equal(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}))
		Expect(err).NotTo(HaveOccurred())
	}),
	Entry("Extensions", "apps/v1", "Deployment", func(gvr schema.GroupVersionResource, err error) {
		Expect(gvr).To(Equal(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}))
		Expect(err).NotTo(HaveOccurred())
	}),
	Entry("CRD", "admissionregistration.k8s.io/v1beta1", "ValidatingWebhookConfiguration", func(gvr schema.GroupVersionResource, err error) {
		Expect(gvr).To(Equal(schema.GroupVersionResource{
			Group:    "admissionregistration.k8s.io",
			Version:  "v1beta1",
			Resource: "validatingwebhookconfigurations",
		}))
		Expect(err).NotTo(HaveOccurred())
	}),
	Entry("Invalid", "a/b/c", "foo", func(gvr schema.GroupVersionResource, err error) {
		Expect(gvr).To(Equal(schema.GroupVersionResource{}))
		Expect(err).To(HaveOccurred())
	}),
)

var _ = DescribeTable("GVKToTypeMeta", func(gvk schema.GroupVersionKind, expected metav1.TypeMeta) {
	Expect(GVKToTypeMeta(gvk)).To(Equal(expected))
},
	Entry("Core", schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"}),
	Entry("Extensions", schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"}),
)
