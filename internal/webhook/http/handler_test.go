package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
)

var _ = Describe("Handler", func() {
	var (
		handler  *Handler
		req      *http.Request
		recorder *httptest.ResponseRecorder
		mgr      *testenv.Manager
	)

	newRequest := func(body *Body) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).To(Succeed())
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")

		return req
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handler = NewHandler(mgr)

		Expect(mgr.Initialize()).To(Succeed())
	})

	JustBeforeEach(func() {
		recorder = httptest.NewRecorder()
		httputil.NewHandler(handler.Handle)(recorder, req)
	})

	AfterEach(func() {
		mgr.Stop()
	})

	for name, body := range map[string]*Body{
		"payload is invalid":        nil,
		"payload without namespace": {},
		"payload without name":      {Namespace: "a"},
		"payload without action":    {Namespace: "a", Name: "b"},
		"invalid action":            {Namespace: "a", Name: "b", Action: "c"},
	} {
		body := body

		When(name, func() {
			BeforeEach(func() {
				req = newRequest(body)
			})

			It("should respond 400", func() {
				Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
			})

			It("should respond errors", func() {
				var res httputil.Response
				Expect(json.NewDecoder(recorder.Body).Decode(&res)).To(Succeed())
				Expect(res.Errors).NotTo(BeEmpty())
			})
		})
	}

	When("no matching webhooks", func() {
		BeforeEach(func() {
			req = newRequest(&Body{
				Namespace: "a",
				Name:      "b",
				Action:    hookutil.ActionApply,
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
				Errors: []httputil.Error{
					{Description: "HTTPWebhook not found"},
				},
			})))
		})
	})
})
