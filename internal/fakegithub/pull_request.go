package fakegithub

import (
	"github.com/google/go-github/v32/github"
	"k8s.io/utils/pointer"
)

type PullRequestEventModifier func(event *github.PullRequestEvent)

func NewPullRequestEvent(modifiers ...PullRequestEventModifier) *github.PullRequestEvent {
	prNumber := 46
	event := &github.PullRequestEvent{
		Number: &prNumber,
		Action: pointer.StringPtr("opened"),
		Repo: &github.Repository{
			Name:     pointer.StringPtr("bar"),
			FullName: pointer.StringPtr("foo/bar"),
			Owner:    &github.User{Login: pointer.StringPtr("foo")},
		},
		PullRequest: &github.PullRequest{
			Base: &github.PullRequestBranch{
				Ref: pointer.StringPtr("base"),
				SHA: pointer.StringPtr("b436f6eb3356504235c0c9a8e74605c820d8d9cc"),
			},
			Head: &github.PullRequestBranch{
				Ref: pointer.StringPtr("test"),
				SHA: pointer.StringPtr("0ce4cf0450de14c6555c563fa9d36be67e69aa2f"),
			},
		},
	}

	for _, mod := range modifiers {
		mod(event)
	}

	return event
}

func SetPullRequestEventAction(action string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.Action = pointer.StringPtr(action)
	}
}

func SetPullRequestEventBranch(branch string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.PullRequest.Base.Ref = pointer.StringPtr(branch)
	}
}

func SetPullRequestEventLabels(labels []string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.PullRequest.Labels = make([]*github.Label, len(labels))

		for i, label := range labels {
			event.PullRequest.Labels[i] = NewLabel(SetLabelName(label))
		}
	}
}

func SetPullRequestEventTriggeredLabel(label string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.Label = NewLabel(SetLabelName(label))
	}
}
