package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("Func", func() {
	var (
		output  interface{}
		reducer Interface
		err     error
	)

	input := 46

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("Success", func() {
		BeforeEach(func() {
			reducer = Func(func(input interface{}) (interface{}, error) {
				return input.(int) * 2, nil
			})
		})

		It("should return reduced result", func() {
			Expect(output).To(Equal(46 * 2))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("func returns an error", func() {
		reduceErr := xerrors.New("reduce err")

		BeforeEach(func() {
			reducer = Func(func(_ interface{}) (interface{}, error) {
				return nil, reduceErr
			})
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, reduceErr)).To(BeTrue())
		})
	})
})

var _ = Describe("Reducers", func() {
	var (
		output  interface{}
		reducer Interface
		err     error
	)

	input := 1

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("Success", func() {
		BeforeEach(func() {
			reducer = Reducers{
				Func(func(input interface{}) (interface{}, error) { return input.(int) * 2, nil }),
				Func(func(input interface{}) (interface{}, error) { return input.(int) + 3, nil }),
			}
		})

		It("should return reduced result", func() {
			Expect(output).To(Equal(1*2 + 3))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("func returns an error", func() {
		reduceErr := xerrors.New("reduce err")

		BeforeEach(func() {
			reducer = Reducers{
				Func(func(input interface{}) (interface{}, error) { return input.(int) * 2, nil }),
				Func(func(input interface{}) (interface{}, error) { return nil, reduceErr }),
				Func(func(input interface{}) (interface{}, error) { return input.(int) + 3, nil }),
			}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, reduceErr)).To(BeTrue())
		})
	})
})
