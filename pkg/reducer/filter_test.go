package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("Filter", func() {
	var (
		input, output interface{}
		filter        Filter
		err           error
	)

	JustBeforeEach(func() {
		output, err = filter.Reduce(input)
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
		filterErr := xerrors.New("filter error")

		BeforeEach(func() {
			filter = Filter{Func: func(_, _, _ interface{}) (bool, error) {
				return false, filterErr
			}}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, filterErr)).To(BeTrue())
		})
	}

	When("type = array", func() {
		BeforeEach(func() {
			input = []string{"abc", "def", "ghi"}
		})

		When("success", func() {
			BeforeEach(func() {
				filter = Filter{Func: func(_, key, _ interface{}) (bool, error) {
					return key.(int)%2 == 0, nil
				}}
			})

			expectSuccess([]interface{}{"abc", "ghi"})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]int{
				"ab": 1,
				"bc": 2,
				"ac": 3,
			}
		})

		When("success", func() {
			BeforeEach(func() {
				filter = Filter{Func: func(value, _, _ interface{}) (bool, error) {
					return value.(int)%2 != 0, nil
				}}
			})

			expectSuccess(map[string]interface{}{
				"ab": 1,
				"ac": 3,
			})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("other types", func() {
		BeforeEach(func() {
			input = nil
			filter = Filter{Func: func(value, _, _ interface{}) (bool, error) {
				return true, nil
			}}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
})
