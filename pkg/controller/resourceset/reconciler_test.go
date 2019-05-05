package resourceset

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/internal/yaml"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func isIntegrationTest() bool {
	_, ok := env.(*testenv.Integration)
	return ok
}

func skipIntegrationTest(msg string) {
	if isIntegrationTest() {
		Skip(msg)
	}
}

var _ = Describe("Reconciler", func() {
	var (
		eventRecorder *record.FakeRecorder
		reconciler    *Reconciler
		result        reconcile.Result
		client        *testutil.Client
		err           error
		namespaceMap  *testutil.Map
	)

	loadTestData := func(name string) []runtime.Object {
		data, err := yaml.DecodeFile(filepath.Join("testdata", name+".yml"))
		Expect(err).NotTo(HaveOccurred())
		data = testutil.SetRandomNamespace(namespaceMap, data)
		return data
	}

	deleteObjects := func(objects []runtime.Object) {
		for _, obj := range objects {
			Expect(client.Delete(context.Background(), obj)).NotTo(HaveOccurred())
		}
	}

	getResourceSet := func() *v1alpha1.ResourceSet {
		rs := new(v1alpha1.ResourceSet)
		Expect(client.Get(context.Background(), types.NamespacedName{
			Namespace: namespaceMap.Value("test"),
			Name:      "test-46",
		}, rs)).NotTo(HaveOccurred())
		return rs
	}

	createAppliedObject := func(object runtime.Object) {
		if object.GetObjectKind().GroupVersionKind().Empty() {
			if gvk, err := apiutil.GVKForObject(object, testutil.Scheme); err == nil {
				object.GetObjectKind().SetGroupVersionKind(gvk)
			}
		}

		rs := getResourceSet()

		if obj, err := meta.Accessor(object); err == nil {
			obj.SetNamespace(rs.Namespace)
			obj.SetName(rs.Name)
			obj.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion:         "pullup.dev/v1alpha1",
					Kind:               "ResourceSet",
					Name:               rs.Name,
					UID:                rs.UID,
					Controller:         pointer.BoolPtr(true),
					BlockOwnerDeletion: pointer.BoolPtr(true),
				},
			})
		}

		Expect(client.Create(context.Background(), object)).NotTo(HaveOccurred())
		client.Reset()
	}

	testSuccess := func(name string) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
			client = testutil.NewClient(env, data...)
			eventRecorder = record.NewFakeRecorder(len(data))
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

		It("should match the golden file", func() {
			skipIntegrationTest("Skip golden file")
			changed := golden.PrepareObjects(testutil.RestoreNamespace(namespaceMap, client.GetChangedObjects()))
			Expect(changed).To(golden.Match(golden.Path(name)))
		})
	}

	BeforeEach(func() {
		reconciler = &Reconciler{
			logger: log.NullLogger{},
		}
		namespaceMap = testutil.NewMap()
	})

	JustBeforeEach(func() {
		reconciler.client = client
		reconciler.recorder = eventRecorder
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-46",
				Namespace: namespaceMap.Value("test"),
			},
		})
	})

	When("resource set does not exist", func() {
		BeforeEach(func() {
			client = testutil.NewClient(env)
			eventRecorder = record.NewFakeRecorder(0)
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("should not record any events", func() {
			Consistently(eventRecorder.Events).ShouldNot(Receive())
		})
	})

	When("original resource exists", func() {
		testSuccess("original-resource-exists")

		It("should record Created event", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Created Created resource v1 Pod: \"test-46\""))
		})
	})

	When("original resource contains read-only properties", func() {
		BeforeEach(func() {
			skipIntegrationTest("Skipped")
		})

		testSuccess("clean-original-resource")
	})

	When("kind = Service", func() {
		testSuccess("service")
	})

	When("applied resource exists", func() {
		testSuccess("applied-resource-exists")

		BeforeEach(func() {
			createAppliedObject(&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			})
		})

		It("should record Updated event", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Updated Updated resource v1 Pod: \"test-46\""))
		})
	})

	When("neither original nor applied resource exists", func() {
		testSuccess("without-original-and-applied")
	})

	When("resource is not controlled", func() {
		BeforeEach(func() {
			data := loadTestData("resource-not-controlled")
			client = testutil.NewClient(env, data...)
			eventRecorder = record.NewFakeRecorder(len(data))
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("should record ResourceExists event", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Warning ResourceExists resource already exists and is not managed by pullup: v1 Pod: \"test-46\""))
		})
	})

	When("resources contain common arrays", func() {
		testSuccess("common-array")

		BeforeEach(func() {
			createAppliedObject(&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
									Args:  []string{"a2", "b2"},
								},
							},
						},
					},
				},
			})
		})
	})

	When("resources contain named arrays", func() {
		testSuccess("named-array")

		BeforeEach(func() {
			createAppliedObject(&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "foo",
									Image: "gcr.io/test/foo:v1",
									Env: []corev1.EnvVar{
										{Name: "B", Value: "b1"},
										{Name: "A", Value: "a2"},
									},
								},
							},
						},
					},
				},
			})
		})
	})

	When("resources contain template", func() {
		testSuccess("template")
	})

	When("resource set contains multiple resources", func() {
		testSuccess("multi-resources")

		It("should record multiple events", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Created Created resource v1 Pod: \"test-46\""))
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Created Created resource v1 Service: \"test-46\""))
		})
	})

	When("resources are not changed", func() {
		testSuccess("unchanged-resources")

		BeforeEach(func() {
			createAppliedObject(&corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			})
		})

		It("should record Unchanged event", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Unchanged Skipped resource v1 Pod: \"test-46\""))
		})

		It("should not change resources", func() {
			Expect(client.GetChangedObjects()).To(BeEmpty())
		})
	})
})
