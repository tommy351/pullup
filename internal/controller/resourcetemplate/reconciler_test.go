package resourcetemplate

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	testError := func(name string, requeue bool) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It(fmt.Sprintf("should return requeue = %v", requeue), func() {
			Expect(result).To(Equal(reconcile.Result{Requeue: requeue}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
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
				Name:      "foo-rt",
				Namespace: namespaceMap.GetRandom("test"),
			},
		})
	})

	When("resource template does not exist", func() {
		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
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
		testSuccess("resource-not-controlled")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeWarning,
			Reason:  ReasonResourceExists,
			Message: "Resource already exists and is not managed by pullup: v1/Pod foo-rt",
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

	When("targetName is given", func() {
		testSuccess("target-name")
		testGolden()
	})

	When("jsonPatch is invalid", func() {
		testError("json-patch-invalid", false)
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeWarning,
			Reason:  ReasonInvalidPatch,
			Message: "replace operation does not apply: doc is missing path: /metadata/annotations/foo: missing value",
		})
	})

	When("metadata is a template string", func() {
		testSuccess("metadata-template")
		testGolden()
	})

	When("multi patches", func() {
		testSuccess("multi-patches")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: "Created resource: v1/Pod foo-rt",
		}, testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: "Created resource: v1/ConfigMap foo-rt",
		})
	})

	When("using CRD", func() {
		testSuccess("crd")
		testGolden()
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonCreated,
			Message: "Created resource: test.pullup.dev/v1/Job foo-rt",
		})
	})

	When("resources are removed from patches", func() {
		testSuccess("delete-resources")
		testEvent(testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonDeleted,
			Message: "Deleted resource: v1/ConfigMap abc",
		}, testenv.EventData{
			Type:    corev1.EventTypeNormal,
			Reason:  ReasonDeleted,
			Message: "Deleted resource: v1/ConfigMap xyz",
		})

		It("should delete resources", func() {
			changes := testenv.GetChanges(reconciler.Client)
			gvk := schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}
			Expect(changes).To(ContainElements(
				testenv.Change{
					Type:             "delete",
					GroupVersionKind: gvk,
					NamespacedName: types.NamespacedName{
						Namespace: namespaceMap.GetRandom("test"),
						Name:      "abc",
					},
				},
				testenv.Change{
					Type:             "delete",
					GroupVersionKind: gvk,
					NamespacedName: types.NamespacedName{
						Namespace: namespaceMap.GetRandom("test"),
						Name:      "xyz",
					},
				},
			))
		})

		BeforeEach(func() {
			client := reconciler.Client
			ctx := context.Background()
			rt := new(v1beta1.ResourceTemplate)
			name := types.NamespacedName{
				Name:      "foo-rt",
				Namespace: namespaceMap.GetRandom("test"),
			}
			err := client.Get(ctx, name, rt)
			Expect(err).NotTo(HaveOccurred())

			now := metav1.Now()
			rt.Status.LastUpdateTime = &now
			rt.Status.Active = []v1beta1.ResourceReference{
				{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "abc",
				},
				{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "def",
				},
				{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "xyz",
				},
			}

			Expect(client.Status().Update(ctx, rt)).To(Succeed())

			By("Wait for status updated")
			Eventually(func() *metav1.Time {
				rt := new(v1beta1.ResourceTemplate)
				Expect(client.Get(ctx, name, rt)).NotTo(HaveOccurred())

				return rt.Status.LastUpdateTime
			}).ShouldNot(BeNil())
		})
	})
})
