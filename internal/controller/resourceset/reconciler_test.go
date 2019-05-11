package resourceset

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
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

	getResourceSet := func() (*v1alpha1.ResourceSet, error) {
		rs := new(v1alpha1.ResourceSet)
		err := testenv.GetClient().Get(context.Background(), types.NamespacedName{
			Namespace: namespaceMap.GetRandom("test"),
			Name:      "test-46",
		}, rs)

		if err != nil {
			return nil, err
		}

		return rs, err
	}

	setOwnerReferences := func(input runtime.Object) {
		obj, err := meta.Accessor(input)

		if err != nil {
			return
		}

		refs := obj.GetOwnerReferences()

		if len(refs) == 0 {
			return
		}

		rs, err := getResourceSet()

		if err != nil {
			return
		}

		for i, ref := range refs {
			if ref.APIVersion == "pullup.dev/v1alpha1" && ref.Kind == "ResourceSet" && ref.Name == rs.Name {
				refs[i].UID = rs.UID
			}
		}
	}

	loadTestData := func(name string) []runtime.Object {
		data, err := testutil.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data = testutil.MapObjects(data, namespaceMap.SetObject)

		for _, obj := range data {
			setOwnerReferences(obj)
			Expect(testenv.GetClient().Create(context.Background(), obj)).To(Succeed())
		}

		return data
	}

	deleteObjects := func(objects []runtime.Object) {
		for _, obj := range objects {
			Expect(testenv.GetClient().Delete(context.Background(), obj)).To(Succeed())
		}
	}

	testSuccess := func(name string) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
		})

		AfterEach(func() {
			deleteObjects(data)
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	}

	testGoldenFile := func(name string) {
		It("should match the golden file", func() {
			events := testenv.GetChanges(reconciler.client)
			objects, err := testenv.GetChangedObjects(events)
			Expect(err).NotTo(HaveOccurred())
			objects = testutil.MapObjects(objects, namespaceMap.RestoreObject)
			Expect(objects).To(golden.MatchObject(fmt.Sprintf("testdata/%s.golden", name)))
		})
	}

	testEvent := func(expected ...testenv.EventData) {
		It("should record event", func() {
			mgr.WaitForSync()

			events, err := testenv.ListEvents()
			Expect(err).NotTo(HaveOccurred())

			for _, e := range expected {
				Expect(testenv.MapEventData(events)).To(ContainElement(e))
			}
		})
	}

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
			Expect(testenv.GetChanges(reconciler.client)).To(BeEmpty())
		})
	})

	When("original resource exists", func() {
		name := "original-resource-exists"

		testSuccess(name)
		testGoldenFile(name)
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
		testGoldenFile(name)
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonUpdated,
			Message: `Updated resource v1 Pod: "test-46"`,
		})
	})

	When("neither original nor applied resource exists", func() {
		name := "without-original-and-applied"

		testSuccess(name)
		testGoldenFile(name)
	})

	When("resource is not controlled", func() {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData("resource-not-controlled")
		})

		AfterEach(func() {
			deleteObjects(data)
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
		testGoldenFile(name)
	})

	When("resource contain named arrays", func() {
		name := "named-array"

		testSuccess(name)
		testGoldenFile(name)
	})

	When("resource contain template", func() {
		name := "template"

		testSuccess(name)
		testGoldenFile(name)
	})

	When("resource set contains multiple resources", func() {
		name := "multi-resources"

		testSuccess(name)
		testGoldenFile(name)
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
			Expect(testenv.GetChanges(reconciler.client)).To(BeEmpty())
		})
	})
})
