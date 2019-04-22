package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("ReduceNested", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]int{
						"c": 1,
					},
				},
			}
		})

		When("success", func() {
			BeforeEach(func() {
				reducer = ReduceNested([]string{"a", "b", "c"}, Func(func(value interface{}) (interface{}, error) {
					return value.(int) * 10, nil
				}))
			})

			It("should return reduced result", func() {
				Expect(output).To(Equal(map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]int{
							"c": 10,
						},
					},
				}))
			})

			It("should have no errors", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("key does not exist", func() {
			BeforeEach(func() {
				reducer = ReduceNested([]string{"a", "b", "d"}, Func(func(value interface{}) (interface{}, error) {
					return 2, nil
				}))
			})

			It("should return reduced result", func() {
				Expect(output).To(Equal(map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]int{
							"c": 1,
							"d": 2,
						},
					},
				}))
			})

			It("should not have errors", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("reducer returns an error", func() {
			reduceErr := xerrors.New("reduce error")

			BeforeEach(func() {
				reducer = ReduceNested([]string{"a", "b", "c"}, Func(func(value interface{}) (interface{}, error) {
					return nil, reduceErr
				}))
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

	When("other types", func() {
		BeforeEach(func() {
			input = map[string]interface{}{
				"a": map[string]interface{}{
					"b": 1,
				},
			}
			reducer = ReduceNested([]string{"a", "b", "c"}, Func(func(value interface{}) (interface{}, error) {
				return value, nil
			}))
		})

		It("should return nil", func() {
			Expect(output).To(BeNil())
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(xerrors.Is(err, ErrNotMap)).To(BeTrue())
		})
	})
})

var _ = Describe("SetNested", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]int{
						"c": 1,
					},
				},
			}
			reducer = SetNested([]string{"a", "b", "c"}, 3)
		})

		It("should delete nested key", func() {
			Expect(output).To(Equal(map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]int{
						"c": 3,
					},
				},
			}))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("DeleteNested", func() {
	var (
		input, output interface{}
		reducer       Interface
		err           error
	)

	JustBeforeEach(func() {
		output, err = reducer.Reduce(input)
	})

	When("type = map", func() {
		BeforeEach(func() {
			input = map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]int{
						"c": 1,
					},
				},
				"d": map[string]int{
					"e": 2,
				},
			}
			reducer = DeleteNested([]string{"a", "b", "c"})
		})

		It("should delete nested key", func() {
			Expect(output).To(Equal(map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]int{},
				},
				"d": map[string]int{
					"e": 2,
				},
			}))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("keys length = 1", func() {
		BeforeEach(func() {
			input = map[string]interface{}{
				"a": 1,
				"d": 2,
			}
			reducer = DeleteNested([]string{"a"})
		})

		It("should delete nested key", func() {
			Expect(output).To(Equal(map[string]interface{}{
				"d": 2,
			}))
		})

		It("should have no errors", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
