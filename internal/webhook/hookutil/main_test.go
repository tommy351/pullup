package hookutil

import (
	"testing"

	"github.com/tommy351/pullup/internal/testenv"
)

func Test(t *testing.T) {
	testenv.RunSpecsInEnvironment(t, "hookutil")
}
