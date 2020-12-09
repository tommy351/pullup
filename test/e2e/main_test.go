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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nolint: gochecknoglobals
var (
	webhookHost  = os.Getenv("WEBHOOK_SERVICE_NAME")
	k8sNamespace = os.Getenv("KUBERNETES_NAMESPACE")
	backoff      = wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Jitter:   0.1,
		Steps:    10,
	}

	scheme    *runtime.Scheme
	k8sClient client.Client
)

func Test(t *testing.T) {
	SetDefaultEventuallyTimeout(time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)
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
	}).Should(And(
		HaveHTTPStatus(http.StatusOK),
		testutil.HaveHTTPHeader("X-Resource-Name", name),
	))
}

func waitUntilObjectDeleted(key types.NamespacedName, obj runtime.Object) {
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := k8sClient.Get(context.TODO(), key, obj)
		if err != nil {
			return true, client.IgnoreNotFound(err)
		}

		return false, nil
	})
	Expect(err).NotTo(HaveOccurred())
}

func getObject(key types.NamespacedName, obj runtime.Object) {
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := k8sClient.Get(context.TODO(), key, obj)
		if err != nil {
			return false, client.IgnoreNotFound(err)
		}

		return true, nil
	})
	Expect(err).NotTo(HaveOccurred())
}

func waitUntilConfigMapDeleted(name string) {
	waitUntilObjectDeleted(types.NamespacedName{
		Namespace: k8sNamespace,
		Name:      name,
	}, &corev1.ConfigMap{})
}

func getConfigMap(name string) *corev1.ConfigMap {
	conf := new(corev1.ConfigMap)
	getObject(types.NamespacedName{
		Namespace: k8sNamespace,
		Name:      name,
	}, conf)

	return conf
}
