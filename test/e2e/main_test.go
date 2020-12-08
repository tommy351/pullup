package e2e

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nolint: gochecknoglobals
var (
	webhookHost  = os.Getenv("WEBHOOK_SERVICE_NAME")
	webhookName  = os.Getenv("WEBHOOK_NAME")
	k8sNamespace = os.Getenv("KUBERNETES_NAMESPACE")

	scheme    *runtime.Scheme
	k8sClient client.Client
)

func Test(t *testing.T) {
	testutil.RunSpecs(t, "e2e")
}

var _ = BeforeSuite(func() {
	config, err := k8s.LoadConfig(k8s.Config{
		Namespace: k8sNamespace,
	})
	Expect(err).NotTo(HaveOccurred())

	scheme, err = k8s.NewScheme()
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(config, client.Options{
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred())
})
