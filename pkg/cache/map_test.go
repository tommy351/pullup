package cache

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Map", func() {
	Describe("Load", func() {
		var m *Map

		BeforeEach(func() {
			m = NewMap()
			m.Store("foo", "bar")
		})

		It("should return (value, true) if exist", func() {
			v, ok := m.Load("foo")
			Expect(v).To(Equal("bar"))
			Expect(ok).To(BeTrue())
		})

		It("should return (nil, false) if not exist", func() {
			v, ok := m.Load("bar")
			Expect(v).To(BeNil())
			Expect(ok).To(BeFalse())
		})
	})

	Describe("Store", func() {
		It("should update the map", func() {
			m := NewMap()
			m.Store("foo", "bar")
			_, ok := m.Load("foo")
			Expect(ok).To(BeTrue())
		})
	})

	Describe("LoadOrStore", func() {
		var m *Map

		BeforeEach(func() {
			m = NewMap()
			m.Store(1, "a")
		})

		Describe("when the key exist", func() {
			var (
				returnValue interface{}
				returnOk    bool
			)

			BeforeEach(func() {
				returnValue, returnOk = m.LoadOrStore(1, func() interface{} {
					return "b"
				})
			})

			It("should return (value, true)", func() {
				Expect(returnValue).To(Equal("a"))
				Expect(returnOk).To(BeTrue())
			})

			It("should not update the map", func() {
				v, ok := m.Load(1)
				Expect(v).To(Equal("a"))
				Expect(ok).To(BeTrue())
			})
		})

		Describe("when the key does not exist", func() {
			var (
				returnValue interface{}
				returnOk    bool
			)

			BeforeEach(func() {
				returnValue, returnOk = m.LoadOrStore(2, func() interface{} {
					return "b"
				})
			})

			It("should return (value, false)", func() {
				Expect(returnValue).To(Equal("b"))
				Expect(returnOk).To(BeFalse())
			})

			It("should update the map", func() {
				v, ok := m.Load(2)
				Expect(v).To(Equal("b"))
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("Delete", func() {
		It("should delete the key from map", func() {
			m := NewMap()
			m.Store("foo", "bar")
			m.Delete("foo")
			_, ok := m.Load("foo")
			Expect(ok).To(BeFalse())
		})
	})
})
