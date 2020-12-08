package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/fakegithub"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Webhook", func() {
	var objects []runtime.Object

	webhookName := "http-server"

	BeforeEach(func() {
		objects = loadObjects("testdata/alpha-webhook.yml")
		createObjects(objects)
	})

	AfterEach(func() {
		deleteObjects(objects)
	})

	When("action = opened", func() {
		event := fakegithub.NewPullRequestEvent()
		name := fmt.Sprintf("%s-%d", webhookName, event.GetNumber())

		BeforeEach(func() {
			Eventually(func() *http.Response {
				var buf bytes.Buffer
				Expect(json.NewEncoder(&buf).Encode(&event)).To(Succeed())

				req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, fmt.Sprintf("http://%s/webhooks/github", webhookHost), &buf)
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-GitHub-Delivery", uuid.Must(uuid.NewRandom()).String())
				req.Header.Set("X-GitHub-Event", "pull_request")

				res, _ := http.DefaultClient.Do(req)

				return res
			}, time.Minute, time.Second).Should(And(
				Not(BeNil()),
				HaveHTTPStatus(http.StatusOK),
			))
		})

		AfterEach(func() {
			rs := &v1alpha1.ResourceSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: k8sNamespace,
					Name:      name,
				},
			}
			Expect(k8sClient.Delete(context.TODO(), rs)).To(Succeed())
		})

		It("should create a service", func() {
			testHTTPServer(name)
		})
	})
})
