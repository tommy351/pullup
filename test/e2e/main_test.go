package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nolint: gochecknoglobals
var (
	webhookHost = os.Getenv("WEBHOOK_SERVICE_NAME")

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

func loadObjects(path string) []runtime.Object {
	objects, err := k8s.LoadObjects(scheme, path)
	Expect(err).NotTo(HaveOccurred())

	return objects
}

func createObjects(objects []runtime.Object) {
	for _, obj := range objects {
		Expect(k8sClient.Create(context.TODO(), obj)).To(Succeed())
	}
}

func deleteObjects(objects []runtime.Object) {
	for _, obj := range objects {
		Expect(client.IgnoreNotFound(k8sClient.Delete(context.TODO(), obj))).To(Succeed())
	}
}

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return http.DefaultClient.Do(req)
}

func testHTTPServer(name string) {
	Eventually(func() (*http.Response, error) {
		return httpGet(fmt.Sprintf("http://%s", name))
	}, time.Minute, time.Second).Should(And(
		HaveHTTPStatus(http.StatusOK),
		testutil.HaveHTTPHeader("X-Resource-Name", name),
	))
}

func checkServiceDeleted(name string) {
	Eventually(func() bool {
		err := k8sClient.Get(context.TODO(), types.NamespacedName{
			Namespace: k8sNamespace,
			Name:      name,
		}, &corev1.Service{})

		return errors.IsNotFound(err)
	}, time.Minute, time.Second).Should(BeTrue())
}
