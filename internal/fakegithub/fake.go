package fakegithub

import (
	"github.com/google/go-github/v32/github"
	"k8s.io/utils/pointer"
)

type PushEventModifier func(event *github.PushEvent)

type PullRequestEventModifier func(event *github.PullRequestEvent)

func NewPushEvent(modifiers ...PushEventModifier) *github.PushEvent {
	event := &github.PushEvent{
		Head: pointer.StringPtr("0ce4cf0450de14c6555c563fa9d36be67e69aa2f"),
		Ref:  pointer.StringPtr("refs/heads/test"),
		Repo: &github.PushEventRepository{
			Name:     pointer.StringPtr("bar"),
			FullName: pointer.StringPtr("foo/bar"),
			Owner: &github.User{
				Login: pointer.StringPtr("foo"),
			},
		},
	}

	for _, mod := range modifiers {
		mod(event)
	}

	return event
}

func SetPushBranch(branch string) PushEventModifier {
	return func(event *github.PushEvent) {
		event.Ref = pointer.StringPtr("refs/heads/" + branch)
	}
}

func SetPushTag(tag string) PushEventModifier {
	return func(event *github.PushEvent) {
		event.Ref = pointer.StringPtr("refs/tags/" + tag)
	}
}

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

func SetPullRequestAction(action string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.Action = pointer.StringPtr(action)
	}
}

func SetPullRequestBranch(branch string) PullRequestEventModifier {
	return func(event *github.PullRequestEvent) {
		event.PullRequest.Base.Ref = pointer.StringPtr(branch)
	}
}
