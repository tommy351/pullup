package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("Handler", func() {
	var (
		handler      *Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
		mgr          *testenv.Manager
		namespaceMap *random.NamespaceMap
	)

	newRequest := func(body *Body) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).To(Succeed())
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")

		ctx := logr.NewContext(req.Context(), log.Log)

		return req.WithContext(ctx)
	}

	loadTestData := func(name string) []runtime.Object {
		data, err := k8s.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data, err = k8s.MapObjects(data, func(obj runtime.Object) error {
			namespaceMap.SetObject(obj)

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(testenv.CreateObjects(data)).To(Succeed())

		return data
	}

	getChanges := func() []testenv.Change {
		return testenv.GetChanges(handler.Client)
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handler = NewHandler(mgr)

		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()
	})

	testGolden := func() {
		It("should match the golden file", func() {
			objects, err := testenv.GetChangedObjects(getChanges())
			Expect(err).NotTo(HaveOccurred())

			objects, err = k8s.MapObjects(objects, func(obj runtime.Object) error {
				namespaceMap.RestoreObject(obj)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(golden.MatchObject())
		})
	}

	testSuccess := func(name string) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It("should respond 200", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	}

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

	When("action = apply", func() {
		BeforeEach(func() {
			req = newRequest(&Body{
				Namespace: namespaceMap.GetRandom("test"),
				Name:      "foobar",
				Action:    hookutil.ActionApply,
			})
		})

		When("resource template exists", func() {
			testSuccess("resource-template-exists")
			testGolden()

			It("should record Updated event", func() {
				Expect(mgr.WaitForEvent(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  hookutil.ReasonUpdated,
					Message: "Updated resource template: foobar-rt",
				})).To(BeTrue())
			})
		})

		When("resource template does not exist", func() {
			testSuccess("resource-template-not-exist")
			testGolden()

			It("should record Created event", func() {
				Expect(mgr.WaitForEvent(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  hookutil.ReasonCreated,
					Message: "Created resource template: foobar-rt",
				})).To(BeTrue())
			})
		})
	})

	When("action = delete", func() {
		BeforeEach(func() {
			req = newRequest(&Body{
				Namespace: namespaceMap.GetRandom("test"),
				Name:      "foobar",
				Action:    hookutil.ActionDelete,
			})
		})

		When("resource template exists", func() {
			testSuccess("resource-template-exists")

			It("should delete the resource template", func() {
				rt := new(v1beta1.ResourceTemplate)
				err := handler.Client.Get(context.Background(), types.NamespacedName{
					Namespace: namespaceMap.GetRandom("test"),
					Name:      "foobar-rt",
				}, rt)
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("should record Deleted event", func() {
				Expect(mgr.WaitForEvent(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  hookutil.ReasonDeleted,
					Message: "Deleted resource template: foobar-rt",
				})).To(BeTrue())
			})
		})

		When("resource template not exist", func() {
			testSuccess("resource-template-not-exist")

			It("should record NotExist event", func() {
				Expect(mgr.WaitForEvent(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  hookutil.ReasonNotExist,
					Message: "Resource template does not exist: foobar-rt",
				})).To(BeTrue())
			})
		})
	})

	When("schema is given", func() {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData("schema")
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		When("data matches the schema", func() {
			BeforeEach(func() {
				req = newRequest(&Body{
					Namespace: namespaceMap.GetRandom("test"),
					Name:      "foobar",
					Action:    hookutil.ActionApply,
					Data: extv1.JSON{
						Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"foo": "bar",
							"bar": 123,
						}),
					},
				})
			})

			testGolden()

			It("should respond 200", func() {
				Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
			})
		})

		When("data does not match the schema", func() {
			BeforeEach(func() {
				req = newRequest(&Body{
					Namespace: namespaceMap.GetRandom("test"),
					Name:      "foobar",
					Action:    hookutil.ActionApply,
					Data: extv1.JSON{
						Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"foo": true,
						}),
					},
				})
			})

			It("should respond 400", func() {
				Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
			})

			It("should respond errors", func() {
				Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
					Errors: []httputil.Error{
						{
							Type:        "invalid_type",
							Description: "Invalid type. Expected: string, given: boolean",
							Field:       "foo",
						},
					},
				})))
			})
		})

		When("data is nil", func() {
			BeforeEach(func() {
				req = newRequest(&Body{
					Namespace: namespaceMap.GetRandom("test"),
					Name:      "foobar",
					Action:    hookutil.ActionApply,
				})
			})

			It("should respond 400", func() {
				Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
			})

			It("should respond errors", func() {
				Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
					Errors: []httputil.Error{
						{
							Type:        "invalid_type",
							Description: "Invalid type. Expected: object, given: null",
							Field:       "(root)",
						},
					},
				})))
			})
		})
	})

	When("schema is invalid", func() {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData("schema-invalid")
			req = newRequest(&Body{
				Namespace: namespaceMap.GetRandom("test"),
				Name:      "foobar",
				Action:    hookutil.ActionApply,
			})
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("should respond errors", func() {
			Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
				Errors: []httputil.Error{
					{Description: "Failed to validate against JSON schema"},
				},
			})))
		})
	})

	When("secretToken is given", func() {
		var data []runtime.Object

		BeforeEach(func() {
			req = newRequest(&Body{
				Namespace: namespaceMap.GetRandom("test"),
				Name:      "foobar",
				Action:    hookutil.ActionApply,
			})
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		When("secret exists", func() {
			BeforeEach(func() {
				data = loadTestData("secret")
				req.Header.Set("Pullup-Webhook-Secret", "some-thing-very-secret")
			})

			It("should respond 200", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})

		When("secret mismatch", func() {
			BeforeEach(func() {
				data = loadTestData("secret")
				req.Header.Set("Pullup-Webhook-Secret", "foobar")
			})

			It("should respond 403", func() {
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})

			It("should respond errors", func() {
				Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
					Errors: []httputil.Error{
						{Description: "Secret mismatch"},
					},
				})))
			})
		})

		When("key does not exist", func() {
			BeforeEach(func() {
				data = loadTestData("secret-key-not-exist")
			})

			It("should respond 403", func() {
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})

			It("should respond errors", func() {
				Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
					Errors: []httputil.Error{
						{Description: "Key does not contain in the secret"},
					},
				})))
			})
		})

		When("secret does not exist", func() {
			BeforeEach(func() {
				data = loadTestData("secret-not-exist")
			})

			It("should respond 403", func() {
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})

			It("should respond errors", func() {
				Expect(recorder.Body.Bytes()).To(MatchJSON(testutil.MustMarshalJSON(&httputil.Response{
					Errors: []httputil.Error{
						{Description: "Secret not found"},
					},
				})))
			})
		})
	})
})
