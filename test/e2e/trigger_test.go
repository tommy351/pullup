package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Trigger", func() {
	var objects []runtime.Object
	triggerName := "conf-ab"

	BeforeEach(func() {
		objects = loadObjects("testdata/trigger.yml")
		createObjects(objects)

		// Wait until resources are applied
		getConfigMap("conf-a-a1b2c3")
		getConfigMap("conf-b-a1b2c3")
	})

	AfterEach(func() {
		deleteObjects(objects)
	})

	When("patches are updated", func() {
		BeforeEach(func() {
			trigger := new(v1beta1.Trigger)
			trigger.Namespace = k8sNamespace
			trigger.Name = triggerName

			err := k8sClient.Patch(context.TODO(), trigger, client.RawPatch(types.JSONPatchType, testutil.MustMarshalJSON([]v1beta1.JSONPatch{
				{
					Operation: v1beta1.JSONPatchOpAdd,
					Path:      "/spec/patches/0/merge",
					Value: &extv1.JSON{
						Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"data": map[string]interface{}{
								"foo": "bar",
							},
						}),
					},
				},
			})))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update ResourceTemplate as well", func() {
			Eventually(func() (map[string]string, error) {
				conf := new(corev1.ConfigMap)
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: k8sNamespace,
					Name:      "conf-a-a1b2c3",
				}, conf)
				if err != nil {
					return nil, fmt.Errorf("failed to get configmap: %w", err)
				}

				return conf.Data, nil
			}).Should(Equal(map[string]string{
				"a":   "abc",
				"foo": "bar",
			}))
		})
	})

	When("patches are removed", func() {
		BeforeEach(func() {
			trigger := new(v1beta1.Trigger)
			trigger.Namespace = k8sNamespace
			trigger.Name = triggerName

			err := k8sClient.Patch(context.TODO(), trigger, client.RawPatch(types.JSONPatchType, testutil.MustMarshalJSON([]v1beta1.JSONPatch{
				{
					Operation: v1beta1.JSONPatchOpRemove,
					Path:      "/spec/patches/1",
				},
			})))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should removed inactive resources", func() {
			waitUntilConfigMapDeleted("conf-b-a1b2c3")
		})
	})
})
