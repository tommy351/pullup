package collection

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NamedArray", func() {
	var arr *NamedArray
	input := []interface{}{
		map[string]interface{}{"name": "a", "value": 1},
		map[string]interface{}{"name": "b", "value": 2},
		map[string]interface{}{"name": "c", "value": 3},
		map[string]interface{}{"name": "d", "value": 4},
	}

	BeforeEach(func() {
		arr, _ = NewNamedArray(input)
	})

	Context("Index", func() {
		When("index exists", func() {
			It("should return the value", func() {
				for i := 0; i < 4; i++ {
					actual, ok := arr.Index(i)
					Expect(ok).To(BeTrue())
					Expect(actual).To(Equal(input[i]))
				}
			})
		})

		When("index does not exist", func() {
			It("should return nil", func() {
				v, ok := arr.Index(4)
				Expect(ok).To(BeFalse())
				Expect(v).To(BeNil())
			})
		})
	})

	Context("Get", func() {
		When("key exists", func() {
			It("should return the value", func() {
				tests := map[string]int{
					"a": 1,
					"b": 2,
					"c": 3,
					"d": 4,
				}

				for k, v := range tests {
					actual, ok := arr.Get(k)
					Expect(ok).To(BeTrue())
					Expect(actual).To(Equal(map[string]interface{}{"name": k, "value": v}))
				}
			})
		})

		When("key does not exist", func() {
			It("should return nil", func() {
				v, ok := arr.Get("z")
				Expect(ok).To(BeFalse())
				Expect(v).To(BeNil())
			})
		})
	})

	Context("Set", func() {
		When("key exists", func() {
			It("should override existing value", func() {
				arr.Set("c", map[string]interface{}{"name": "c", "value": 10})
				Expect(arr.ToArray()).To(Equal([]interface{}{
					map[string]interface{}{"name": "a", "value": 1},
					map[string]interface{}{"name": "b", "value": 2},
					map[string]interface{}{"name": "c", "value": 10},
					map[string]interface{}{"name": "d", "value": 4},
				}))
			})
		})

		When("key does not exist", func() {
			It("should append the value", func() {
				value := map[string]interface{}{"name": "e", "value": 10}
				arr.Set("e", value)
				expected := append(input, value)
				Expect(arr.ToArray()).To(Equal(expected))
			})
		})
	})

	Context("Len", func() {
		It("should return length of array", func() {
			Expect(arr.Len()).To(Equal(4))
		})
	})

	Context("Iterate", func() {
		It("should iterate over the array", func() {
			var keys []string
			var values []interface{}
			iter := arr.Iterate()

			for iter.Next() {
				keys = append(keys, iter.Key())
				values = append(values, iter.Value())
			}

			Expect(keys).To(Equal([]string{"a", "b", "c", "d"}))
			Expect(values).To(Equal(input))
		})
	})

	Context("ToArray", func() {
		It("should return the raw array", func() {
			Expect(arr.ToArray()).To(Equal(input))
		})
	})
})

var _ = Describe("NewNamedArray", func() {
	var (
		data interface{}
		arr  *NamedArray
		ok   bool
	)

	expectError := func() {
		It("should return nil", func() {
			Expect(arr).To(BeNil())
		})

		It("should return false", func() {
			Expect(ok).To(BeFalse())
		})
	}

	JustBeforeEach(func() {
		arr, ok = NewNamedArray(data)
	})

	When("type = array", func() {
		When("all elements have a name", func() {
			BeforeEach(func() {
				data = []interface{}{
					map[string]interface{}{"name": "a", "value": 1},
					map[string]interface{}{"name": "b", "value": 2},
					map[string]interface{}{"name": "c", "value": 3},
					map[string]interface{}{"name": "d", "value": 4},
				}
			})

			It("should not return nil", func() {
				Expect(arr).NotTo(BeNil())
			})

			It("should return true", func() {
				Expect(ok).To(BeTrue())
			})
		})

		When("not all elements have a name", func() {
			BeforeEach(func() {
				data = []interface{}{
					map[string]interface{}{"name": "a"},
					map[string]interface{}{"foo": "b"},
				}
			})

			expectError()
		})

		When("element type = non-map", func() {
			BeforeEach(func() {
				data = []interface{}{1, 2, 3}
			})

			expectError()
		})
	})

	When("other types", func() {
		BeforeEach(func() {
			data = "foo"
		})

		expectError()
	})
})
