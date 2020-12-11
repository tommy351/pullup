package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
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
