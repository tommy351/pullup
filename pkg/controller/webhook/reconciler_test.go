package webhook

import (
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
				Name:      "bar",
				Namespace: "foo",
			},
		})
	})

	When("webhook does not exist", func() {
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

	When("patch success", func() {
		BeforeEach(func() {
			data := loadTestData("success")
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
			Expect(client.Changed).To(testutil.MatchGolden("testdata/success.golden"))
		})

		It("should record Patched events", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Patched Patched resource set \"bar-46\""))
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Patched Patched resource set \"bar-64\""))
		})
	})
})
