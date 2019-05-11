package merge

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Any", func() {
	var (
		input, output, source interface{}
		err                   error
	)

	expectOutputTo := func(matcher types.GomegaMatcher) {
		It("should return expected value", func() {
			Expect(output).To(matcher)
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	JustBeforeEach(func() {
		output, err = Any(input, source)
	})

	When("input type = array", func() {
		BeforeEach(func() {
			input = []string{"a", "b", "c"}
		})

		When("source type = array", func() {
			When("same element types", func() {
				When("source length > input length", func() {
					BeforeEach(func() {
						source = []string{"q", "w", "o", "p"}
					})

					expectOutputTo(Equal([]string{"q", "w", "o", "p"}))
				})

				When("source length < input length", func() {
					BeforeEach(func() {
						source = []string{"x", "y"}
					})

					expectOutputTo(Equal([]string{"x", "y", "c"}))
				})
			})

			When("different element types", func() {
				BeforeEach(func() {
					source = []int{1, 2}
				})

				expectOutputTo(Equal([]interface{}{1, 2, "c"}))
			})

			When("nested array", func() {
				BeforeEach(func() {
					input = []interface{}{
						[]int{1, 2},
						[]uint{5, 6},
					}
					source = []interface{}{
						[]int{11},
						[]uint{15},
					}
				})

				expectOutputTo(Equal([]interface{}{
					[]int{11, 2},
					[]uint{15, 6},
				}))
			})
		})

		When("source type = non-array", func() {
			BeforeEach(func() {
				source = "aaa"
			})

			expectOutputTo(Equal("aaa"))
		})

		When("source is nil", func() {
			BeforeEach(func() {
				source = nil
			})

			expectOutputTo(BeNil())
		})
	})

	When("input type = map", func() {
		BeforeEach(func() {
			input = map[string]string{
				"a": "d",
				"b": "e",
				"c": "f",
			}
		})

		When("source type = map", func() {
			When("same types", func() {
				BeforeEach(func() {
					source = map[string]string{
						"a": "x",
						"d": "y",
					}
				})

				expectOutputTo(Equal(map[string]string{
					"a": "x",
					"b": "e",
					"c": "f",
					"d": "y",
				}))
			})

			When("different types", func() {
				BeforeEach(func() {
					source = map[int]int{
						1: 2,
						3: 4,
					}
				})

				expectOutputTo(Equal(map[interface{}]interface{}{
					"a": "d",
					"b": "e",
					"c": "f",
					1:   2,
					3:   4,
				}))
			})
		})

		When("source type = non-map", func() {
			BeforeEach(func() {
				source = "bbb"
			})

			expectOutputTo(Equal("bbb"))
		})

		When("source is nil", func() {
			BeforeEach(func() {
				source = nil
			})

			expectOutputTo(BeNil())
		})

		When("nested map", func() {
			BeforeEach(func() {
				input = map[string]interface{}{
					"a": map[string]int{"x": 1, "y": 2},
					"b": map[int]string{1: "x", 2: "y"},
				}
				source = map[string]interface{}{
					"a": map[string]int{"x": 3},
					"b": map[int]string{1: "z"},
				}
			})

			expectOutputTo(Equal(map[string]interface{}{
				"a": map[string]int{"x": 3, "y": 2},
				"b": map[int]string{1: "z", 2: "y"},
			}))
		})
	})

	When("other input types", func() {
		BeforeEach(func() {
			input = "ccc"
			source = "bar"
		})

		expectOutputTo(Equal("bar"))
	})
})
