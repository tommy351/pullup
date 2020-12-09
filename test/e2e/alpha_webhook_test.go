package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/tommy351/pullup/internal/fakegithub"
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
			sendGitHubRequest("pull_request", event)
		})

		It("should create a service", func() {
			testHTTPServer(name)
		})
	})
})
