package k8s

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("StringP", func() {
	It("should return a pointer", func() {
		Expect(StringP("test")).To(PointTo(Equal("test")))
	})
})

var _ = Describe("IntP", func() {
	It("should return a pointer", func() {
		Expect(IntP(46)).To(PointTo(Equal(46)))
	})
})

var _ = Describe("BoolP", func() {
	It("should return a pointer", func() {
		Expect(BoolP(true)).To(PointTo(Equal(true)))
	})
})
