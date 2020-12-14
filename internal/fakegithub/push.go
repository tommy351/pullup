package fakegithub

import (
	"github.com/google/go-github/v32/github"
	"k8s.io/utils/pointer"
)

type PushEventModifier func(event *github.PushEvent)

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

func SetPushEventBranch(branch string) PushEventModifier {
	return func(event *github.PushEvent) {
		event.Ref = pointer.StringPtr("refs/heads/" + branch)
	}
}

func SetPushEventTag(tag string) PushEventModifier {
	return func(event *github.PushEvent) {
		event.Ref = pointer.StringPtr("refs/tags/" + tag)
	}
}
