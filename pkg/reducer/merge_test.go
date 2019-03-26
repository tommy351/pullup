package reducer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"
)

var _ = Describe("Merge", func() {
	var (
		input, output, source interface{}
		err                   error
	)

	JustBeforeEach(func() {
		output, err = Merge{Source: source}.Reduce(input)
	})

	When("type = []interface{}", func() {
		BeforeEach(func() {
			source = []interface{}{"x", "y", "z"}
		})

		JustAfterEach(func() {
			Expect(err).NotTo(HaveOccurred())
		})

		When("input is nil", func() {
			BeforeEach(func() {
				input = nil
			})

			It("should return source", func() {
				Expect(output).To(Equal(source))
			})
		})

		When("input is not nil", func() {
			BeforeEach(func() {
				input = []interface{}{"a", "b", "c"}
			})

			It("should replace with source", func() {
				Expect(output).To(Equal([]interface{}{"x", "y", "z"}))
			})
		})

		When("it is a named array", func() {
			BeforeEach(func() {
				input = []interface{}{
					map[string]interface{}{"name": "a", "a": 1, "b": 2},
					map[string]interface{}{"name": "b", "a": 3, "b": 4},
				}
				source = []interface{}{
					map[string]interface{}{"name": "a", "a": 5},
					map[string]interface{}{"name": "b", "b": 6},
				}
			})

			It("should merge by names", func() {
				Expect(output).To(ConsistOf([]interface{}{
					map[string]interface{}{"name": "a", "a": 5, "b": 2},
					map[string]interface{}{"name": "b", "a": 3, "b": 6},
				}))
			})
		})
	})

	When("type = map[string]interface{}", func() {
		BeforeEach(func() {
			source = map[string]interface{}{"a": "foo", "b": "bar"}
		})

		JustAfterEach(func() {
			Expect(err).NotTo(HaveOccurred())
		})

		When("input is nil", func() {
			BeforeEach(func() {
				input = nil
			})

			It("should return source", func() {
				Expect(output).To(Equal(source))
			})
		})

		When("input is not nil", func() {
			BeforeEach(func() {
				input = map[string]interface{}{"a": "baz", "c": "boo"}
			})

			It("should merge by keys", func() {
				Expect(output).To(Equal(map[string]interface{}{
					"a": "foo",
					"b": "bar",
					"c": "boo",
				}))
			})
		})
	})

	When("type = reducer.Interface", func() {
		When("input is nil", func() {
			BeforeEach(func() {
				input = nil
				source = Map{}
			})

			It("should return nil", func() {
				Expect(output).To(BeNil())
			})

			It("should not have error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("input is not nil", func() {
			BeforeEach(func() {
				input = map[string]interface{}{"a": 1, "b": 2}
			})

			When("success", func() {
				BeforeEach(func() {
					source = Map{Func: func(_, key, _ interface{}) (interface{}, error) {
						return key, nil
					}}
				})

				It("should execute the reducer", func() {
					Expect(output).To(Equal(map[string]interface{}{"a": "a", "b": "b"}))
				})

				It("should not have error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("reducer returns an error", func() {
				reduceErr := xerrors.New("reducer error")

				BeforeEach(func() {
					source = Map{Func: func(_, key, _ interface{}) (interface{}, error) {
						return nil, reduceErr
					}}
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
	})
})
