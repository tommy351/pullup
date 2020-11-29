package resourcetemplate

import (
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	testGolden := func() {
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

		reconciler = NewReconciler(mgr, logr.Discard())
		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()
	})

	AfterEach(func() {
		mgr.Stop()
	})

	JustBeforeEach(func() {
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "foo-rt",
				Namespace: namespaceMap.GetRandom("test"),
			},
		})
	})

	When("resource template does not exist", func() {
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
		testSuccess("original-resource-exists")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: "Created resource: v1/Pod foo-rt",
		})
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
			Message: `resource already exists and is not managed by pullup: v1/Pod foo-rt`,
		})
	})

	When("merge is given", func() {
		testSuccess("merge")
		testGolden()
	})

	When("merge with apiVersion, kind and name set", func() {
		testSuccess("merge-with-gvk-and-name")
		testGolden()
	})

	When("merge without data", func() {
		testSuccess("merge-without-data")
		testGolden()
	})

	When("jsonPatch is given", func() {
		testSuccess("json-patch")
		testGolden()
	})

	When("original and current resource exists", func() {
		testSuccess("original-and-current-resource-exists")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonPatched,
			Message: "Patched resource: v1/Pod foo-rt",
		})
	})

	When("resource is unchanged", func() {
		testSuccess("unchanged-resource")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonUnchanged,
			Message: "Skipped resource: v1/Pod foo-rt",
		})
	})

	When("without original and current resource", func() {
		testSuccess("without-original-and-current-resource")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: "Created resource: v1/Pod foo-rt",
		})
	})

	When("kind = Service", func() {
		testSuccess("service")
	})
})
