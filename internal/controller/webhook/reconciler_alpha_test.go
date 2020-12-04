package webhook

import (
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

var _ = Describe("AlphaReconciler", func() {
	var (
		reconciler   *AlphaReconciler
		mgr          *testenv.Manager
		result       reconcile.Result
		err          error
		namespaceMap *random.NamespaceMap
	)

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		reconciler = NewAlphaReconciler(mgr, log.Log)
		Expect(mgr.Initialize()).To(Succeed())

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
			Expect(testenv.GetChanges(reconciler.Client)).To(BeEmpty())
		})
	})

	// nolint: dupl
	When("patch success", func() {
		var data []runtime.Object

		BeforeEach(func() {
			var err error
			data, err = k8s.LoadObjects(testenv.GetScheme(), "testdata/success.yml")
			Expect(err).NotTo(HaveOccurred())

			data, err = k8s.MapObjects(data, func(obj runtime.Object) error {
				namespaceMap.SetObject(obj)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
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

		It("should record patched events", func() {
			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonPatched,
				Message: `Patched resource set "bar-46"`,
			})).To(BeTrue())

			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonPatched,
				Message: `Patched resource set "bar-64"`,
			})).To(BeTrue())
		})
	})
})