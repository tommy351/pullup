package webhook

import (
	"testing"

	"github.com/tommy351/pullup/internal/testutil"
)

func Test(t *testing.T) {
	testutil.RunSpecs(t, "webhook")
}
