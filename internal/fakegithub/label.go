package fakegithub

import (
	"fmt"

	"github.com/google/go-github/v32/github"
	"k8s.io/utils/pointer"
)

type LabelModifier func(label *github.Label)

func NewLabel(modifiers ...LabelModifier) *github.Label {
	label := &github.Label{
		Color:   pointer.StringPtr("123456"),
		Default: pointer.BoolPtr(false),
	}

	SetLabelName("hotfix")

	for _, mod := range modifiers {
		mod(label)
	}

	return label
}

func SetLabelName(name string) LabelModifier {
	return func(label *github.Label) {
		label.Name = pointer.StringPtr(name)
		label.URL = pointer.StringPtr(fmt.Sprintf("https://github.com/foo/bar/labels/%s", name))
		label.Description = pointer.StringPtr(fmt.Sprintf("Description for %s", name))
	}
}
