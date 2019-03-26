package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"golang.org/x/xerrors"
)

var _ = Describe("Unique", func() {
	var (
		input, output interface{}
		uniq          Unique
		err           error
	)

	JustBeforeEach(func() {
		output, err = uniq.Reduce(input)
	})

	expectSuccess := func(matcher types.GomegaMatcher) {
		It("should return filtered result", func() {
			Expect(output).To(matcher)
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	expectFuncError := func() {
		uniqErr := xerrors.New("unique error")

		BeforeEach(func() {
			uniq = Unique{Func: func(_, _, _ interface{}) (interface{}, error) {
				return nil, uniqErr
			}}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, uniqErr)).To(BeTrue())
		})
	}

	When("type = array", func() {
		BeforeEach(func() {
			input = []string{"ab", "bc", "ac"}
		})

		When("success", func() {
			BeforeEach(func() {
				uniq = Unique{Func: func(value, _, _ interface{}) (interface{}, error) {
					return value.(string)[0], nil
				}}
			})

			expectSuccess(Equal([]interface{}{"ab", "bc"}))
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]int{"ab": 1, "bc": 2, "ac": 3}
		})

		When("success", func() {
			BeforeEach(func() {
				uniq = Unique{Func: func(_, key, _ interface{}) (interface{}, error) {
					return key.(string)[0], nil
				}}
			})

			expectSuccess(Or(
				Equal(map[string]interface{}{"ab": 1, "bc": 2}),
				Equal(map[string]interface{}{"bc": 2, "ac": 3}),
			))
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("other types", func() {
		BeforeEach(func() {
			input = nil
			uniq = Unique{Func: func(value, _, _ interface{}) (interface{}, error) {
				return value, nil
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
