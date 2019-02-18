package kubernetes

import (
	"fmt"

	"github.com/tommy351/pullup/pkg/config"
)

type Resource struct {
	config.ResourceConfig

	PullRequestNumber int
	HeadCommitSHA     string
}

func (r *Resource) ModifiedName() string {
	return fmt.Sprintf("%s-pr%d", r.Name, r.PullRequestNumber)
}
