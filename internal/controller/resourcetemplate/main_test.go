package resourcetemplate

import (
	"testing"

	"github.com/tommy351/pullup/internal/testenv"
)

func Test(t *testing.T) {
	testenv.RunSpecsInEnvironment(t, "controller/resourcetemplate")
}
