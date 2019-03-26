package reducer

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("Map", func() {
	var (
		input, output interface{}
		mapper        Map
		err           error
	)

	JustBeforeEach(func() {
		output, err = mapper.Reduce(input)
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
		mapErr := xerrors.New("map error")

		BeforeEach(func() {
			mapper = Map{Func: func(_, _, _ interface{}) (interface{}, error) {
				return nil, mapErr
			}}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, mapErr)).To(BeTrue())
		})
	}

	When("type = array", func() {
		BeforeEach(func() {
			input = []string{"abc", "def", "ghi"}
		})

		When("success", func() {
			BeforeEach(func() {
				mapper = Map{Func: func(value, key, _ interface{}) (interface{}, error) {
					return fmt.Sprintf("%v%v", key, value), nil
				}}
			})

			expectSuccess([]interface{}{"0abc", "1def", "2ghi"})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]string{
				"a": "bc",
				"d": "ef",
				"g": "hi",
			}
		})

		When("Success", func() {
			BeforeEach(func() {
				mapper = Map{Func: func(value, key, _ interface{}) (interface{}, error) {
					return fmt.Sprintf("%v%v", key, value), nil
				}}
			})

			expectSuccess(map[string]interface{}{
				"a": "abc",
				"d": "def",
				"g": "ghi",
			})
		})

		When("func returns an error", func() {
			expectFuncError()
		})
	})

	When("other types", func() {
		BeforeEach(func() {
			input = nil
			mapper = Map{Func: func(value, _, _ interface{}) (interface{}, error) {
				return value, nil
			}}
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return an error", func() {
			Expect(xerrors.Is(err, ErrNotCollection)).To(BeTrue())
		})
	})
})
