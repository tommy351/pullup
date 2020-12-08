package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/fakegithub"
	"k8s.io/apimachinery/pkg/runtime"
)

func sendGitHubRequest(event string, data interface{}) {
	Eventually(func() (*http.Response, error) {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(&data)).To(Succeed())

		req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, fmt.Sprintf("http://%s/webhooks/github", webhookHost), &buf)
		Expect(err).NotTo(HaveOccurred())

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Delivery", uuid.Must(uuid.NewRandom()).String())
		req.Header.Set("X-GitHub-Event", event)

		return http.DefaultClient.Do(req)
	}).Should(HaveHTTPStatus(http.StatusOK))
}

var _ = Describe("GitHubWebhook", func() {
	var objects []runtime.Object

	When("event = push", func() {
		webhookName := "http-server-push"
		event := fakegithub.NewPushEvent()
		name := fmt.Sprintf("%s-%s", webhookName, event.GetHead())

		BeforeEach(func() {
			objects = loadObjects("testdata/github-webhook-push.yml")
			createObjects(objects)
			sendGitHubRequest("push", event)
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		It("should create a service", func() {
			testHTTPServer(name)
		})
	})

	When("event = pull_request", func() {
		webhookName := "http-server-pull-request"

		BeforeEach(func() {
			objects = loadObjects("testdata/github-webhook-pull-request.yml")
			createObjects(objects)
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		When("action = opened", func() {
			event := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestAction("opened"))
			name := fmt.Sprintf("%s-%d", webhookName, event.GetNumber())

			BeforeEach(func() {
				sendGitHubRequest("pull_request", event)
			})

			It("should create a service", func() {
				testHTTPServer(name)
			})
		})

		When("action = closed", func() {
			openedEvent := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestAction("opened"))
			closedEvent := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestAction("closed"))
			name := fmt.Sprintf("%s-%d", webhookName, openedEvent.GetNumber())

			BeforeEach(func() {
				sendGitHubRequest("pull_request", openedEvent)
				testHTTPServer(name)
				sendGitHubRequest("pull_request", closedEvent)
			})

			It("should delete the service", func() {
				checkServiceDeleted(name)
			})
		})
	})
})
