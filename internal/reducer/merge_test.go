package reducer

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Merge", func() {
	It("works", func() {
		reducer := Merge(map[string]interface{}{
			"a": map[string]interface{}{
				"b": 1,
				"d": 5,
			},
			"e": 6,
		})

		actual, err := reducer.Reduce(map[string]interface{}{
			"a": map[string]interface{}{
				"b": 2,
				"c": 3,
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(map[string]interface{}{
			"a": map[string]interface{}{
				"b": 1,
				"c": 3,
				"d": 5,
			},
			"e": 6,
		}))
	})
})

var _ = Describe("MergeWith", func() {
	var (
		output  interface{}
		reducer Interface
		err     error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(2)
	})

	When("Success", func() {
		BeforeEach(func() {
			reducer = MergeWith(3, func(input, source interface{}) (interface{}, error) {
				return input.(int) * source.(int), nil
			})
		})

		It("should return the result", func() {
			Expect(output).To(Equal(6))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("func returns an error", func() {
		mergeErr := errors.New("merge err")

		BeforeEach(func() {
			reducer = MergeWith(nil, func(input, source interface{}) (interface{}, error) {
				return nil, mergeErr
			})
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should have no errors", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, mergeErr)).To(BeTrue())
		})
	})
})
