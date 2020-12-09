package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HTTPWebhook", func() {
	var (
		data        map[string]interface{}
		webhookName string
	)

	sendRequest := func(action string, headers map[string]string) {
		Eventually(func() (*http.Response, error) {
			var buf bytes.Buffer
			Expect(json.NewEncoder(&buf).Encode(map[string]interface{}{
				"namespace": k8sNamespace,
				"name":      webhookName,
				"action":    action,
				"data":      data,
			})).To(Succeed())

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, fmt.Sprintf("http://%s/webhooks/http", webhookHost), &buf)
			Expect(err).NotTo(HaveOccurred())

			req.Header.Set("Content-Type", "application/json")

			for k, v := range headers {
				req.Header.Set(k, v)
			}

			return http.DefaultClient.Do(req)
		}).Should(And(
			Not(BeNil()),
			HaveHTTPStatus(http.StatusOK),
		))
	}

	Context("configmap", func() {
		var (
			objects []runtime.Object
			suffix  string
		)

		waitUntilApplyDone := func() {
			sendRequest("apply", nil)
			getConfigMap("conf-a-" + suffix)
			getConfigMap("conf-b-" + suffix)
		}

		BeforeEach(func() {
			objects = loadObjects("testdata/http-webhook-config.yml")
			suffix = rand.String(5)
			webhookName = "conf-ab"
			data = map[string]interface{}{
				"suffix": suffix,
				"a":      rand.String(5),
				"b":      rand.String(5),
			}
			createObjects(objects)
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		When("action = apply", func() {
			BeforeEach(func() {
				sendRequest("apply", nil)
			})

			It("should create resources", func() {
				confA := getConfigMap("conf-a-" + suffix)
				confB := getConfigMap("conf-b-" + suffix)

				Expect(confA.Data).To(Equal(map[string]string{
					"a": data["a"].(string),
				}))
				Expect(confB.Data).To(Equal(map[string]string{
					"b": data["b"].(string),
				}))
			})
		})

		When("action = delete", func() {
			BeforeEach(func() {
				waitUntilApplyDone()
				sendRequest("delete", nil)
			})

			It("should delete resources", func() {
				waitUntilConfigMapDeleted("conf-a-" + suffix)
				waitUntilConfigMapDeleted("conf-b-" + suffix)
			})
		})

		When("webhook is updated", func() {
			BeforeEach(func() {
				waitUntilApplyDone()

				webhook := new(v1beta1.HTTPWebhook)
				webhook.Namespace = k8sNamespace
				webhook.Name = webhookName

				err := k8sClient.Patch(context.TODO(), webhook, client.RawPatch(types.JSONPatchType, testutil.MustMarshalJSON([]v1beta1.JSONPatch{
					{
						Operation: "add",
						Path:      "/spec/patches/0/merge/data/foo",
						Value: &extv1.JSON{
							Raw: testutil.MustMarshalJSON("bar"),
						},
					},
				})))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should update ResourceTemplate as well", func() {
				Eventually(func() (map[string]string, error) {
					conf := new(corev1.ConfigMap)
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: k8sNamespace,
						Name:      "conf-a-" + suffix,
					}, conf)
					if err != nil {
						return nil, fmt.Errorf("failed to get configmap: %w", err)
					}

					return conf.Data, nil
				}).Should(Equal(map[string]string{
					"a":   data["a"].(string),
					"foo": "bar",
				}))
			})
		})

		When("patches are removed from the webhook", func() {
			BeforeEach(func() {
				waitUntilApplyDone()

				webhook := new(v1beta1.HTTPWebhook)
				webhook.Namespace = k8sNamespace
				webhook.Name = webhookName

				err := k8sClient.Patch(context.TODO(), webhook, client.RawPatch(types.JSONPatchType, testutil.MustMarshalJSON([]v1beta1.JSONPatch{
					{
						Operation: "remove",
						Path:      "/spec/patches/1",
					},
				})))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should remove inactive resources", func() {
				waitUntilConfigMapDeleted("conf-b-" + suffix)
			})
		})
	})

	Context("service", func() {
		var (
			objects []runtime.Object
			suffix  string
		)

		BeforeEach(func() {
			objects = loadObjects("testdata/http-webhook-service.yml")
			suffix = rand.String(5)
			webhookName = "http-server"
			data = map[string]interface{}{
				"suffix": suffix,
			}
			createObjects(objects)
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		When("action = apply", func() {
			BeforeEach(func() {
				sendRequest("apply", nil)
			})

			It("should create a service", func() {
				testHTTPServer("http-server-" + suffix)
			})
		})
	})

	When("secretToken is given", func() {
		var objects []runtime.Object

		BeforeEach(func() {
			objects = loadObjects("testdata/http-webhook-secret.yml")
			webhookName = "secret"
			data = nil
			createObjects(objects)
			sendRequest("apply", map[string]string{
				"Pullup-Webhook-Secret": "some-thing-very-secret",
			})
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		It("should create resources", func() {
			conf := getConfigMap("secret-rt")
			Expect(conf.Data).To(Equal(map[string]string{
				"a": "abc",
			}))
		})
	})
})
