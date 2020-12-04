package resourceset

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconciler", func() {
	var (
		reconciler   *Reconciler
		mgr          *testenv.Manager
		result       reconcile.Result
		err          error
		namespaceMap *random.NamespaceMap
	)

	loadTestData := func(name string) []runtime.Object {
		data, err := k8s.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data, err = k8s.MapObjects(data, func(obj runtime.Object) error {
			namespaceMap.SetObject(obj)

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(testenv.CreateObjects(data)).To(Succeed())

		return data
	}

	testSuccess := func(name string) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	testGoldenFile := func() {
		It("should match the golden file", func() {
			changes := testenv.GetChanges(reconciler.Client)
			objects, err := testenv.GetChangedObjects(changes)
			Expect(err).NotTo(HaveOccurred())

			objects, err = k8s.MapObjects(objects, func(obj runtime.Object) error {
				namespaceMap.RestoreObject(obj)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(golden.MatchObject())
		})
	}

	testEvent := func(expected ...testenv.EventData) {
		It("should record event", func() {
			for _, e := range expected {
				Expect(mgr.WaitForEvent(e)).To(BeTrue())
			}
		})
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		reconciler = NewReconciler(mgr, log.Log)
		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()
	})

	AfterEach(func() {
		mgr.Stop()
	})

	JustBeforeEach(func() {
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-46",
				Namespace: namespaceMap.GetRandom("test"),
			},
		})
	})

	When("resource set does not exist", func() {
		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("should not change anything", func() {
			Expect(testenv.GetChanges(reconciler.Client)).To(BeEmpty())
		})
	})

	When("original resource exists", func() {
		name := "original-resource-exists"

		testSuccess(name)
		testGoldenFile()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: `Created resource v1 Pod: "test-46"`,
		})
	})

	When("kind = Service", func() {
		testSuccess("service")
	})

	When("applied resource exists", func() {
		name := "applied-resource-exists"

		testSuccess(name)
		testGoldenFile()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonUpdated,
			Message: `Updated resource v1 Pod: "test-46"`,
		})
	})

	When("neither original nor applied resource exists", func() {
		name := "without-original-and-applied"

		testSuccess(name)
		testGoldenFile()
	})

	When("resource is not controlled", func() {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData("resource-not-controlled")
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})

		testEvent(testenv.EventData{
			Type:    corev1.EventTypeWarning,
			Reason:  ReasonResourceExists,
			Message: `resource already exists and is not managed by pullup: v1 Pod: "test-46"`,
		})
	})

	When("resource contain common arrays", func() {
		name := "common-array"

		testSuccess(name)
		testGoldenFile()
	})

	When("resource contain named arrays", func() {
		name := "named-array"

		testSuccess(name)
		testGoldenFile()
	})

	When("resource contain template", func() {
		name := "template"

		testSuccess(name)
		testGoldenFile()
	})

	When("resource set contains multiple resources", func() {
		name := "multi-resources"

		testSuccess(name)
		testGoldenFile()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: `Created resource v1 Pod: "test-46"`,
		}, testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: `Created resource v1 ConfigMap: "test-46"`,
		})
	})

	When("resources are not changed", func() {
		testSuccess("unchanged-resources")
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonUnchanged,
			Message: `Skipped resource v1 Pod: "test-46"`,
		})

		It("should not change anything", func() {
			Expect(testenv.GetChanges(reconciler.Client)).To(BeEmpty())
		})
	})
})
