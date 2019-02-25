package model

import (
	"fmt"

	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Resource struct {
	config.ResourceConfig

	PullRequestNumber int
	HeadCommitSHA     string
	OriginalResource  *unstructured.Unstructured
	AppliedResource   *unstructured.Unstructured
	PatchedResource   *unstructured.Unstructured
}

func (r *Resource) ModifiedName() string {
	return fmt.Sprintf("%s-pullup-%d", r.Name, r.PullRequestNumber)
}
