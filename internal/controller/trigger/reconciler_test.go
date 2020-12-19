package trigger

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconciler", func() {
	var (
		reconciler   *Reconciler
		mgr          *testenv.Manager
		result       reconcile.Result
		err          error
		namespaceMap *random.NamespaceMap
		conf         ReconcilerConfig
	)

	getChanges := func() []testenv.Change {
		return testenv.GetChanges(reconciler.Client)
	}

	loadTestData := func(name string) []client.Object {
		data, err := k8s.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data, err = k8s.MapObjects(data, namespaceMap.SetObject)
		Expect(err).NotTo(HaveOccurred())
		Expect(testenv.CreateObjects(data)).To(Succeed())

		return data
	}

	testSuccess := func() {
		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	testGolden := func() {
		It("should match the golden file", func() {
			objects, err := testenv.GetChangedObjects(getChanges())
			Expect(err).NotTo(HaveOccurred())

			objects, err = k8s.MapObjects(objects, namespaceMap.RestoreObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(golden.MatchObject())
		})
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		conf = NewReconcilerConfig(mgr)
		reconciler, err = NewReconciler(conf, mgr)
		Expect(err).NotTo(HaveOccurred())
		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()
	})

	AfterEach(func() {
		mgr.Stop()
	})

	JustBeforeEach(func() {
		result, err = reconciler.Reconcile(context.TODO(), reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "bar",
				Namespace: namespaceMap.GetRandom("foo"),
			},
		})
	})

	When("trigger does not exist", func() {
		testSuccess()

		It("should not change anything", func() {
			Expect(getChanges()).To(BeEmpty())
		})
	})

	When("patch success", func() {
		var data []client.Object

		BeforeEach(func() {
			data = loadTestData("success")
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		testSuccess()
		testGolden()

		It("should record patched events", func() {
			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonPatched,
				Message: `Patched resource template: bar-46`,
			})).To(BeTrue())

			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonPatched,
				Message: `Patched resource template: bar-64`,
			})).To(BeTrue())
		})
	})

	When("patches are not changed", func() {
		var data []client.Object

		BeforeEach(func() {
			data = loadTestData("unchanged")
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		testSuccess()

		It("should not update resources", func() {
			Expect(getChanges()).To(BeEmpty())
		})
	})
})
