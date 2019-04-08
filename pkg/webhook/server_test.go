package webhook

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Server", func() {
	var (
		req *http.Request
		res *http.Response
	)

	JustBeforeEach(func() {
		server := &Server{}
		router := server.newRouter()
		router.GET("/panic", func(_ http.ResponseWriter, _ *http.Request) {
			panic("panic test")
		})
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		res = recorder.Result()
	})

	Context("Index", func() {
		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/", nil)
		})

		It("should respond 200", func() {
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("should respond ok", func() {
			Eventually(gbytes.BufferReader(res.Body)).Should(gbytes.Say("ok"))
		})
	})

	Context("Not found", func() {
		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/notfound", nil)
		})

		It("should respond 404", func() {
			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should respond not found", func() {
			Eventually(gbytes.BufferReader(res.Body)).Should(gbytes.Say("Not found"))
		})
	})

	Context("Panic", func() {
		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/panic", nil)
		})

		It("should respond 500", func() {
			Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
		})

		It("should respond internal server error", func() {
			Eventually(gbytes.BufferReader(res.Body)).Should(gbytes.Say("Internal server error"))
		})
	})
})
