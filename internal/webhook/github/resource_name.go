package github

import (
	"github.com/google/go-github/v25/github"
	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
)

const defaultResourceName = "{{ .Webhook.Name }}-{{ .Spec.Number }}"

type nameTemplate struct {
	Webhook *v1alpha1.Webhook
	Spec    *v1alpha1.ResourceSetSpec
	Repo    *github.Repository
}

func getResourceName(event *github.PullRequestEvent, hook *v1alpha1.Webhook, rs *v1alpha1.ResourceSet) (string, error) {
	resourceName := hook.Spec.ResourceName

	if resourceName == "" {
		resourceName = defaultResourceName
	}

	return template.Render(resourceName, &nameTemplate{
		Webhook: hook,
		Spec:    &rs.Spec,
		Repo:    event.Repo,
	})
}
