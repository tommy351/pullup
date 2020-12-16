package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/fakegithub"
	"github.com/tommy351/pullup/internal/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	}).Should(And(
		HaveHTTPStatus(http.StatusOK),
		testutil.HaveHTTPHeader("Content-Type", "application/json"),
	))
}

var _ = Describe("GitHubWebhook", func() {
	var objects []client.Object

	When("event = push", func() {
		webhookName := "conf-push"
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

		It("should create resources", func() {
			conf := getConfigMap(name)
			Expect(conf.Data).To(Equal(map[string]string{
				"a": event.GetHead(),
			}))
		})
	})

	When("event = pull_request", func() {
		webhookName := "conf-pull-request"

		BeforeEach(func() {
			objects = loadObjects("testdata/github-webhook-pull-request.yml")
			createObjects(objects)
		})

		AfterEach(func() {
			deleteObjects(objects)
		})

		When("action = opened", func() {
			event := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened"))
			name := fmt.Sprintf("%s-%d", webhookName, event.GetNumber())

			BeforeEach(func() {
				sendGitHubRequest("pull_request", event)
			})

			It("should create resources", func() {
				conf := getConfigMap(name)
				Expect(conf.Data).To(Equal(map[string]string{
					"a": strconv.Itoa(event.GetNumber()),
				}))
			})
		})

		When("action = closed", func() {
			openedEvent := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened"))
			closedEvent := fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("closed"))
			name := fmt.Sprintf("%s-%d", webhookName, openedEvent.GetNumber())

			BeforeEach(func() {
				sendGitHubRequest("pull_request", openedEvent)
				getConfigMap(name)
				sendGitHubRequest("pull_request", closedEvent)
			})

			It("should delete resources", func() {
				waitUntilConfigMapDeleted(name)
			})
		})
	})
})
