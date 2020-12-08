package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HTTPWebhook", func() {
	var (
		suffix string
		name   string
	)

	webhookName := os.Getenv("ALPHA_WEBHOOK_NAME")

	sendRequest := func(action string) {
		Eventually(func() *http.Response {
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

			res, _ := http.DefaultClient.Do(req)

			return res
		}, time.Minute, time.Second).Should(And(
			Not(BeNil()),
			HaveHTTPStatus(http.StatusOK),
		))
	}

	BeforeEach(func() {
		suffix = rand.String(5)
		name = fmt.Sprintf("%s-%s", webhookName, suffix)
	})

	AfterEach(func() {
		rt := &v1beta1.ResourceTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: k8sNamespace,
				Name:      name,
			},
		}
		Expect(client.IgnoreNotFound(k8sClient.Delete(context.TODO(), rt))).To(Succeed())
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
			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: k8sNamespace,
					Name:      name,
				}, &corev1.Service{})

				return errors.IsNotFound(err)
			}, time.Minute, time.Second).Should(BeTrue())
		})
	})
})
