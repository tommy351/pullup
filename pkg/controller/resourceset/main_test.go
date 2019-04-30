package resourceset

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
)

func Test(t *testing.T) {
	testutil.RunSpecs(t, "resourceset")
}

// nolint: gochecknoglobals
var env testenv.Interface

var _ = BeforeSuite(func() {
	env = testutil.NewEnvironment()
	Expect(env.Start()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(env.Stop()).NotTo(HaveOccurred())
})
