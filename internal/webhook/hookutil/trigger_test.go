package hookutil

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("TriggerHandler", func() {
	var (
		handler      *TriggerHandler
		mgr          *testenv.Manager
		namespaceMap *random.NamespaceMap
		err          error
		options      *TriggerOptions
		webhook      *v1beta1.HTTPWebhook
	)

	loadTestData := func(name string) []runtime.Object {
		data, err := k8s.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data, err = k8s.MapObjects(data, namespaceMap.SetObject)
		Expect(err).NotTo(HaveOccurred())
		Expect(testenv.CreateObjects(data)).To(Succeed())

		return data
	}

	getChanges := func() []testenv.Change {
		return testenv.GetChanges(handler.Client)
	}

	testSuccess := func(name string) {
		var objects []runtime.Object

		BeforeEach(func() {
			objects = loadTestData(name)
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(objects)).To(Succeed())
		})

		It("should not return errors", func() {
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
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handler = NewTriggerHandler(mgr)

		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()

		webhook = &v1beta1.HTTPWebhook{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook",
				Namespace: namespaceMap.GetRandom("test"),
			},
		}
		Expect(testenv.CreateObjects([]runtime.Object{webhook})).To(Succeed())
	})

	JustBeforeEach(func() {
		err = handler.Handle(context.TODO(), options)
	})

	AfterEach(func() {
		Expect(testenv.DeleteObjects([]runtime.Object{webhook})).To(Succeed())
		mgr.Stop()
	})

	When("single trigger", func() {
		testTriggeredEvent := func() {
			It("should record Triggered event", func() {
				Expect(mgr.WaitForEvent(testenv.EventData{
					Type:    corev1.EventTypeNormal,
					Reason:  ReasonTriggered,
					Message: fmt.Sprintf("Triggered: %s/trigger-a", namespaceMap.GetRandom("test")),
				})).To(BeTrue())
			})
		}

		BeforeEach(func() {
			options = &TriggerOptions{
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		When("action = create", func() {
			BeforeEach(func() {
				options.Action = v1beta1.ActionCreate
			})

			When("resource exists", func() {
				testSuccess("resource-exists")
				testGolden()
				testTriggeredEvent()

				It("should record AlreadyExists event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonAlreadyExists,
						Message: "Resource template already exists: trigger-a",
					})).To(BeTrue())
				})
			})

			When("resource does not exist", func() {
				testSuccess("resource-not-exist")
				testGolden()
				testTriggeredEvent()

				It("should record Created event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonCreated,
						Message: "Created resource template: trigger-a",
					})).To(BeTrue())
				})
			})
		})

		When("action = update", func() {
			BeforeEach(func() {
				options.Action = v1beta1.ActionUpdate
			})

			When("resource exists", func() {
				testSuccess("resource-exists")
				testGolden()
				testTriggeredEvent()

				It("should record Updated event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonUpdated,
						Message: "Updated resource template: trigger-a",
					})).To(BeTrue())
				})
			})

			When("resource does not exist", func() {
				testSuccess("resource-not-exist")
				testGolden()
				testTriggeredEvent()

				It("should record NotExist event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonNotExist,
						Message: "Resource template does not exist: trigger-a",
					})).To(BeTrue())
				})
			})
		})

		When("action = apply", func() {
			BeforeEach(func() {
				options.Action = v1beta1.ActionApply
			})

			When("resource exists", func() {
				testSuccess("resource-exists")
				testGolden()
				testTriggeredEvent()

				It("should record Updated event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonUpdated,
						Message: "Updated resource template: trigger-a",
					})).To(BeTrue())
				})
			})

			When("resource does not exist", func() {
				testSuccess("resource-not-exist")
				testGolden()
				testTriggeredEvent()

				It("should record Created event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonCreated,
						Message: "Created resource template: trigger-a",
					})).To(BeTrue())
				})
			})
		})

		When("action = delete", func() {
			BeforeEach(func() {
				options.Action = v1beta1.ActionDelete
			})

			When("resource exists", func() {
				testSuccess("resource-exists")
				testTriggeredEvent()

				It("should delete the resource template", func() {
					rt := new(v1beta1.ResourceTemplate)
					err := handler.Client.Get(context.Background(), types.NamespacedName{
						Namespace: namespaceMap.GetRandom("test"),
						Name:      "trigger-a",
					}, rt)
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				It("should record Deleted event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonDeleted,
						Message: "Deleted resource template: trigger-a",
					})).To(BeTrue())
				})
			})

			When("resource does not exist", func() {
				testSuccess("resource-not-exist")
				testTriggeredEvent()

				It("should record NotExist event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    corev1.EventTypeNormal,
						Reason:  ReasonNotExist,
						Message: "Resource template does not exist: trigger-a",
					})).To(BeTrue())
				})
			})
		})
	})

	When("trigger not found", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-xyz"},
				},
			}
		})

		It("should return the error", func() {
			var tnfe TriggerNotFoundError
			Expect(errors.As(err, &tnfe)).To(BeTrue())
			Expect(tnfe.key).To(Equal(types.NamespacedName{
				Namespace: namespaceMap.GetRandom("test"),
				Name:      "trigger-xyz",
			}))
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})

	When("multiple triggers", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
					{Name: "trigger-b"},
				},
			}
		})

		testSuccess("multiple-triggers")
		testGolden()

		It("should record Created events", func() {
			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonCreated,
				Message: "Created resource template: trigger-a",
			})).To(BeTrue())

			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonCreated,
				Message: "Created resource template: trigger-b",
			})).To(BeTrue())
		})

		It("should record Triggered event", func() {
			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonTriggered,
				Message: fmt.Sprintf("Triggered: %s/trigger-a", namespaceMap.GetRandom("test")),
			})).To(BeTrue())

			Expect(mgr.WaitForEvent(testenv.EventData{
				Type:    corev1.EventTypeNormal,
				Reason:  ReasonTriggered,
				Message: fmt.Sprintf("Triggered: %s/trigger-b", namespaceMap.GetRandom("test")),
			})).To(BeTrue())
		})
	})

	When("trigger in other namespace", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{
						Name:      "foobar",
						Namespace: namespaceMap.GetRandom("test2"),
					},
				},
			}
		})

		testSuccess("trigger-in-other-namespace")
		testGolden()
	})

	When("schema is given", func() {
		var objects []runtime.Object

		BeforeEach(func() {
			objects = loadTestData("schema")
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(objects)).To(Succeed())
		})

		When("data matches the schema", func() {
			BeforeEach(func() {
				options.Event = map[string]interface{}{
					"foo": "bar",
					"bar": 123,
				}
			})

			testGolden()

			It("should not return errors", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("data does not match the schema", func() {
			BeforeEach(func() {
				options.Event = map[string]interface{}{
					"foo": true,
				}
			})

			It("should return errors", func() {
				var ve JSONSchemaValidationErrors
				Expect(errors.As(err, &ve)).To(BeTrue())
			})
		})

		When("data is nil", func() {
			BeforeEach(func() {
				options.Event = nil
			})

			It("should return errors", func() {
				var ve JSONSchemaValidationErrors
				Expect(errors.As(err, &ve)).To(BeTrue())
			})
		})
	})

	When("schema is invalid", func() {
		var objects []runtime.Object

		BeforeEach(func() {
			objects = loadTestData("schema-invalid")
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(objects)).To(Succeed())
		})

		It("should return the error", func() {
			var ve JSONSchemaValidateError
			Expect(errors.As(err, &ve)).To(BeTrue())
		})
	})

	When("EventSourceTrigger.transform is given", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action: v1beta1.ActionApply,
				Source: webhook,
				Event: map[string]interface{}{
					"foo": "abc",
					"bar": "xyz",
				},
				Triggers: []v1beta1.EventSourceTrigger{
					{
						Name: "trigger-a",
						Transform: &extv1.JSON{Raw: testutil.MustMarshalJSON(map[string]interface{}{
							"foo": "{{ .event.bar }}",
							"bar": "{{ .event.foo }}",
						})},
					},
				},
			}
		})

		testSuccess("resource-not-exist")
		testGolden()
	})

	When("action is invalid", func() {
		var objects []runtime.Object

		BeforeEach(func() {
			objects = loadTestData("resource-not-exist")
			options = &TriggerOptions{
				Action: "foo",
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(objects)).To(Succeed())
		})

		It("should return the error", func() {
			Expect(errors.Is(err, ErrInvalidAction)).To(BeTrue())
		})
	})

	When("default action is given", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				DefaultAction: v1beta1.ActionCreate,
				Source:        webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		When("action is empty", func() {
			testSuccess("resource-not-exist")

			It("should create the resource", func() {
				Expect(getChanges()).To(ContainElement(testenv.Change{
					Type: "create",
					NamespacedName: types.NamespacedName{
						Name:      "trigger-a",
						Namespace: namespaceMap.GetRandom("test"),
					},
					GroupVersionKind: v1beta1.GroupVersion.WithKind("ResourceTemplate"),
				}))
			})
		})

		When("action is not empty", func() {
			BeforeEach(func() {
				options.Action = v1beta1.ActionUpdate
			})

			testSuccess("resource-not-exist")

			It("should not have any changes", func() {
				Expect(getChanges()).To(BeEmpty())
			})
		})
	})

	When("action is a template string", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action: "{{ .event.foo }}",
				Source: webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
				Event: map[string]interface{}{
					"foo": "create",
				},
			}
		})

		testSuccess("resource-not-exist")

		It("should create the resource", func() {
			Expect(getChanges()).To(ContainElement(testenv.Change{
				Type: "create",
				NamespacedName: types.NamespacedName{
					Name:      "trigger-a",
					Namespace: namespaceMap.GetRandom("test"),
				},
				GroupVersionKind: v1beta1.GroupVersion.WithKind("ResourceTemplate"),
			}))
		})
	})

	When("access default action in action template", func() {
		BeforeEach(func() {
			options = &TriggerOptions{
				Action:        "{{ .action }}",
				DefaultAction: v1beta1.ActionCreate,
				Source:        webhook,
				Triggers: []v1beta1.EventSourceTrigger{
					{Name: "trigger-a"},
				},
			}
		})

		testSuccess("resource-not-exist")

		It("should create the resource", func() {
			Expect(getChanges()).To(ContainElement(testenv.Change{
				Type: "create",
				NamespacedName: types.NamespacedName{
					Name:      "trigger-a",
					Namespace: namespaceMap.GetRandom("test"),
				},
				GroupVersionKind: v1beta1.GroupVersion.WithKind("ResourceTemplate"),
			}))
		})
	})
})
