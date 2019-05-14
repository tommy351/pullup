package webhook

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
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

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		reconciler = NewReconciler(mgr, log.NullLogger{})
		Expect(mgr.Initialize()).To(Succeed())
		mgr.WaitForSync()

		namespaceMap = random.NewNamespaceMap()
	})

	AfterEach(func() {
		mgr.Stop()
	})

	JustBeforeEach(func() {
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "bar",
				Namespace: namespaceMap.GetRandom("foo"),
			},
		})
	})

	When("webhook does not exist", func() {
		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("should not change anything", func() {
			Expect(testenv.GetChanges(reconciler.client)).To(BeEmpty())
		})
	})

	When("patch success", func() {
		var data []runtime.Object

		BeforeEach(func() {
			var err error
			data, err = testutil.LoadObjects(testenv.GetScheme(), "testdata/success.yml")
			Expect(err).NotTo(HaveOccurred())

			data = testutil.MapObjects(data, namespaceMap.SetObject)
			Expect(testenv.CreateObjects(data)).To(Succeed())
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

		It("should match the golden file", func() {
			changes := testenv.GetChanges(reconciler.client)
			objects, err := testenv.GetChangedObjects(changes)
			Expect(err).NotTo(HaveOccurred())
			objects = testutil.MapObjects(objects, namespaceMap.RestoreObject)
			Expect(objects).To(golden.MatchObject("testdata/success.golden"))
		})

		It("should record patched events", func() {
			mgr.WaitForSync()
			events, err := testenv.ListEvents()
			Expect(err).NotTo(HaveOccurred())

			Expect(testenv.MapEventData(events)).To(SatisfyAll(
				ContainElement(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  ReasonPatched,
					Message: `Patched resource set "bar-46"`,
				}),
				ContainElement(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  ReasonPatched,
					Message: `Patched resource set "bar-64"`,
				}),
			))
		})
	})
})
