package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("DeepMapValue", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

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
			reducer = DeepMapValue(func(_ interface{}) (interface{}, error) {
				return nil, mapErr
			})
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, mapErr)).To(BeTrue())
		})
	}

	BeforeEach(func() {
		input = map[string]interface{}{
			"a": []int{1, 2, 3},
			"b": map[string]interface{}{
				"c": []string{"d", "e"},
			},
		}
	})

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("Success", func() {
		BeforeEach(func() {
			reducer = DeepMapValue(func(value interface{}) (interface{}, error) {
				switch v := value.(type) {
				case int:
					return v * 2, nil
				case map[string]interface{}:
					output := make([]interface{}, 0, len(v))

					for _, mv := range v {
						output = append(output, mv)
					}

					return output, nil
				default:
					return v, nil
				}
			})
		})

		expectSuccess(map[string]interface{}{
			"a": []int{2, 4, 6},
			"b": []interface{}{[]string{"d", "e"}},
		})
	})

	When("func returns an error", func() {
		expectFuncError()
	})
})
