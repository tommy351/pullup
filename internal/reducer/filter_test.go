package reducer

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilterKey", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	expectSuccess := func(expected interface{}) {
		It("should return filtered result", func() {
			Expect(output).To(Equal(expected))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	expectFuncError := func() {
		filterErr := errors.New("filter error")

		BeforeEach(func() {
			reducer = FilterKey(func(_ interface{}) (bool, error) {
				return false, filterErr
			})
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, filterErr)).To(BeTrue())
		})
	}

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			}
		})

		When("Success", func() {
			BeforeEach(func() {
				reducer = FilterKey(func(value interface{}) (bool, error) {
					return value.(string) != "b", nil
				})
			})

			expectSuccess(map[string]int{
				"a": 1,
				"c": 3,
			})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("other types", func() {
		expectError := func() {
			It("should return nil", func() {
				Expect(output).To(BeNil())
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, ErrNotMap)).To(BeTrue())
			})
		}

		BeforeEach(func() {
			input = 5566
			reducer = FilterKey(func(value interface{}) (bool, error) {
				return true, nil
			})
		})

		expectError()
	})
})
