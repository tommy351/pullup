package reducer

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MapValue", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	expectSuccess := func(expected interface{}) {
		It("should return mapped result", func() {
			Expect(output).To(Equal(expected))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	expectFuncError := func() {
		// nolint: goerr113
		mapErr := errors.New("map error")

		BeforeEach(func() {
			reducer = MapValue(func(_ interface{}) (interface{}, error) {
				return nil, mapErr
			})
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, mapErr)).To(BeTrue())
		})
	}

	When("type = array", func() {
		BeforeEach(func() {
			input = []int{1, 2, 3}
		})

		When("Success", func() {
			BeforeEach(func() {
				reducer = MapValue(func(value interface{}) (interface{}, error) {
					return value.(int) * 2, nil
				})
			})

			expectSuccess([]int{2, 4, 6})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

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
				reducer = MapValue(func(value interface{}) (interface{}, error) {
					return value.(int) * 2, nil
				})
			})

			expectSuccess(map[string]int{
				"a": 2,
				"b": 4,
				"c": 6,
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
				Expect(errors.Is(err, ErrNotArrayOrMap)).To(BeTrue())
			})
		}

		BeforeEach(func() {
			input = 5566
			reducer = MapValue(func(value interface{}) (interface{}, error) {
				return value, nil
			})
		})

		expectError()
	})
})
