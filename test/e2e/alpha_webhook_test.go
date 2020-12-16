package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/tommy351/pullup/internal/fakegithub"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Webhook", func() {
	var objects []client.Object

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
			sendGitHubRequest("pull_request", event)
		})

		It("should create a service", func() {
			testHTTPServer(name)
		})
	})
})
