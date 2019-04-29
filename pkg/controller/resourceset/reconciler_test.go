package resourceset

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconciler", func() {
	var (
		eventRecorder *record.FakeRecorder
		reconciler    *Reconciler
		result        reconcile.Result
		client        *testutil.Client
		err           error
	)

	loadTestData := func(name string) []runtime.Object {
		data, err := testutil.DecodeYAMLFile(filepath.Join("testdata", name+".yml"))
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	testSuccess := func(name string) {
		BeforeEach(func() {
			data := loadTestData(name)
			client = testutil.NewClient(data...)
			eventRecorder = record.NewFakeRecorder(len(data))
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should match the golden file", func() {
			Expect(client.Changed).To(testutil.MatchGolden(fmt.Sprintf("testdata/%s.golden", name)))
		})
	}

	BeforeEach(func() {
		reconciler = &Reconciler{
			logger: log.NullLogger{},
		}
	})

	JustBeforeEach(func() {
		reconciler.client = client
		reconciler.EventRecorder = eventRecorder
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-46",
				Namespace: "test",
			},
		})
	})

	When("resource set does not exist", func() {
		BeforeEach(func() {
			client = testutil.NewClient()
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
		testSuccess("clean-original-resource")
	})

	When("kind = Service", func() {
		testSuccess("service")
	})

	When("applied resource exists", func() {
		testSuccess("applied-resource-exists")

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
			client = testutil.NewClient(data...)
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
	})

	When("resources contain named arrays", func() {
		testSuccess("named-array")
	})

	When("resources contain template", func() {
		testSuccess("template")
	})

	When("resource set contains multiple resources", func() {
		testSuccess("multi-resources")

		It("should record multiple events", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Created Created resource v1 Pod: \"test-46\""))
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Updated Updated resource v1 Service: \"test-46\""))
		})
	})
})
