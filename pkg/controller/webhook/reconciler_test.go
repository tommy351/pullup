package webhook

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/internal/yaml"
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
		mapper        *testutil.Map
	)

	loadTestData := func(name string) []runtime.Object {
		data, err := yaml.DecodeFile(filepath.Join("testdata", name+".yml"))
		Expect(err).NotTo(HaveOccurred())
		data = testutil.SetRandomNamespace(mapper, data)
		return data
	}

	BeforeEach(func() {
		reconciler = &Reconciler{
			logger: log.NullLogger{},
		}
		mapper = testutil.NewMap()
	})

	JustBeforeEach(func() {
		reconciler.client = client
		reconciler.EventRecorder = eventRecorder
		result, err = reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "bar",
				Namespace: mapper.Value("foo"),
			},
		})
	})

	When("webhook does not exist", func() {
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

	When("patch success", func() {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData("success")
			client = testutil.NewClient(env, data...)
			eventRecorder = record.NewFakeRecorder(len(data))
		})

		AfterEach(func() {
			for _, obj := range data {
				Expect(client.Delete(context.Background(), obj)).NotTo(HaveOccurred())
			}
		})

		It("should not requeue", func() {
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should not return the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should match the golden file", func() {
			changed := golden.PrepareObjects(testutil.RestoreNamespace(mapper, client.GetChangedObjects()))
			Expect(changed).To(golden.Match(golden.Path("success")))
		})

		It("should record Patched events", func() {
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Patched Patched resource set \"bar-46\""))
			Expect(<-eventRecorder.Events).To(HavePrefix("Normal Patched Patched resource set \"bar-64\""))
		})
	})
})
