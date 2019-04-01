package k8s

import (
	"testing"

	"github.com/tommy351/pullup/internal/testutil"
)

func Test(t *testing.T) {
	testutil.RunSpecs(t, "k8s")
}
