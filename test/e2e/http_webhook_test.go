package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HTTPWebhook", func() {
	var (
		suffix  string
		name    string
		objects []runtime.Object
	)

	webhookName := "http-server"

	sendRequest := func(action string) {
		Eventually(func() (*http.Response, error) {
			var buf bytes.Buffer
			Expect(json.NewEncoder(&buf).Encode(map[string]interface{}{
				"namespace": k8sNamespace,
				"name":      webhookName,
				"action":    action,
				"data": map[string]interface{}{
					"suffix": suffix,
				},
			})).To(Succeed())

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, fmt.Sprintf("http://%s/webhooks/http", webhookHost), &buf)
			Expect(err).NotTo(HaveOccurred())

			req.Header.Set("Content-Type", "application/json")

			return http.DefaultClient.Do(req)
		}, time.Minute, time.Second).Should(And(
			Not(BeNil()),
			HaveHTTPStatus(http.StatusOK),
		))
	}

	BeforeEach(func() {
		suffix = rand.String(5)
		name = fmt.Sprintf("%s-%s", webhookName, suffix)
		objects = loadObjects("testdata/http-webhook.yml")
		createObjects(objects)
	})

	AfterEach(func() {
		deleteObjects(objects)
	})

	When("action = apply", func() {
		BeforeEach(func() {
			sendRequest("apply")
		})

		It("should create a service", func() {
			testHTTPServer(name)
		})
	})

	When("action = delete", func() {
		BeforeEach(func() {
			sendRequest("apply")
			testHTTPServer(name)
			sendRequest("delete")
		})

		It("should delete the service", func() {
			checkServiceDeleted(name)
		})
	})

	When("webhook is updated", func() {
		BeforeEach(func() {
			sendRequest("apply")
			testHTTPServer(name)

			webhook := new(v1beta1.HTTPWebhook)
			webhook.Namespace = k8sNamespace
			webhook.Name = webhookName

			err := k8sClient.Patch(context.TODO(), webhook, client.RawPatch(types.JSONPatchType, testutil.MustMarshalJSON([]v1beta1.JSONPatch{
				{
					Operation: "replace",
					Path:      "/spec/patches/0/merge/spec/template/spec/containers/0/env/0/value",
					Value: &extv1.JSON{
						Raw: testutil.MustMarshalJSON("{{ .webhook.metadata.name }}-{{ .event.suffix }}-new"),
					},
				},
			})))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update the ResourceTemplate as well", func() {
			Eventually(func() (*http.Response, error) {
				return httpGet(fmt.Sprintf("http://%s", name))
			}, time.Minute, time.Second).Should(
				testutil.HaveHTTPHeader("X-Resource-Name", fmt.Sprintf("%s-new", name)),
			)
		})
	})

	When("patches are removed from the webhook", func() {
		BeforeEach(func() {
			sendRequest("apply")
			testHTTPServer(name)

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
			checkServiceDeleted(name)
		})
	})
})
