package github

import (
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

// nolint: gochecknoglobals
var defaultPullRequestTypes = []v1beta1.GitHubPullRequestEventType{
	"opened", "synchronize", "reopened", "closed",
}

func extractRepositoryBeta(hook *v1beta1.GitHubWebhook, name string) *v1beta1.GitHubRepository {
	for _, r := range hook.Spec.Repositories {
		r := r

		if r.Name == name {
			return &r
		}
	}

	return nil
}

func filterByPullRequestType(types []v1beta1.GitHubPullRequestEventType, action string) bool {
	if len(types) == 0 {
		types = defaultPullRequestTypes
	}

	for _, t := range types {
		if string(t) == action {
			return true
		}
	}

	return false
}
