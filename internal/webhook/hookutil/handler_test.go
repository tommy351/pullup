package hookutil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/testutil"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("NewHandler", func() {
	testHandler := func(handler httputil.Handler) *httptest.ResponseRecorder {
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/", nil)
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()
		NewHandler(handler).ServeHTTP(recorder, req)

		return recorder
	}

	When("error is nil", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})
		})

		It("should respond 200", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
		})
	})

	When("error is ErrInvalidAction", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				return ErrInvalidAction
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body).To(MatchJSON(testutil.MustMarshalJSON(httputil.Response{
				Errors: []httputil.Error{
					{Description: "Invalid action"},
				},
			})))
		})
	})

	When("error is JSONSchemaValidationErrors", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				_, err := ValidateJSONSchema(
					&extv1.JSON{
						Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"foo": map[string]interface{}{
									"type": "string",
								},
								"bar": map[string]interface{}{
									"type": "string",
								},
							},
						}),
					},
					&extv1.JSON{
						Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"foo": 3,
							"bar": 4,
						}),
					},
				)

				return err
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			var res httputil.Response
			Expect(json.NewDecoder(recorder.Body).Decode(&res)).To(Succeed())
			Expect(res.Errors).To(ConsistOf([]httputil.Error{
				{Type: "invalid_type", Field: "foo", Description: "Invalid type. Expected: string, given: integer"},
				{Type: "invalid_type", Field: "bar", Description: "Invalid type. Expected: string, given: integer"},
			}))
		})
	})

	When("error is ValidationErrors", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				return ValidationErrors{"err1", "err2"}
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body).To(MatchJSON(testutil.MustMarshalJSON(httputil.Response{
				Errors: []httputil.Error{
					{Description: "err1"},
					{Description: "err2"},
				},
			})))
		})
	})

	When("error is JSONSchemaValidateError", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				return JSONSchemaValidateError{}
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body).To(MatchJSON(testutil.MustMarshalJSON(httputil.Response{
				Errors: []httputil.Error{
					{Description: "Failed to validate against JSON schema"},
				},
			})))
		})
	})

	When("error is TriggerNotFoundError", func() {
		var recorder *httptest.ResponseRecorder

		BeforeEach(func() {
			recorder = testHandler(func(w http.ResponseWriter, r *http.Request) error {
				return TriggerNotFoundError{}
			})
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body).To(MatchJSON(testutil.MustMarshalJSON(httputil.Response{
				Errors: []httputil.Error{
					{Description: "Trigger not found"},
				},
			})))
		})
	})

	When("other errors", func() {
		It("should panic", func() {
			Expect(func() {
				testHandler(func(w http.ResponseWriter, r *http.Request) error {
					// nolint: goerr113
					return errors.New("random err")
				})
			}).To(Panic())
		})
	})
})
