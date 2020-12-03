package resourcetemplate

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testenv"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func Test(t *testing.T) {
	testenv.RunSpecsInEnvironment(t, "controller/resourcetemplate")
}

var _ = BeforeSuite(func() {
	_, err := testenv.Env.InstallCRDs(envtest.CRDInstallOptions{
		Paths: []string{"testdata/crds"},
	})
	Expect(err).NotTo(HaveOccurred())
})
