package manager

import (
	"context"
	"errors"
	"sync"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Handler", func() {
	var handler *Handler

	BeforeEach(func() {
		handler = &Handler{
			Kind:     v1alpha1.Kind("Webhook"),
			MaxRetry: 5,
		}
	})

	Context("OnAdd", func() {
		BeforeEach(func() {
			handler.updateQueue = handler.newQueue("update")
		})

		AfterEach(func() {
			handler.updateQueue.ShutDown()
		})

		expected := &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		}

		JustBeforeEach(func() {
			handler.OnAdd(expected)
		})

		It("should add to the queue", func() {
			actual, _ := handler.updateQueue.Get()
			Expect(actual).To(Equal(expected))
		})
	})

	Context("OnUpdate", func() {
		var newObj interface{}

		oldObj := &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "foo",
				ResourceVersion: "1",
			},
		}

		BeforeEach(func() {
			handler.updateQueue = handler.newQueue("update")
		})

		AfterEach(func() {
			handler.updateQueue.ShutDown()
		})

		JustBeforeEach(func() {
			handler.OnUpdate(oldObj, newObj)
		})

		When("resource version is same", func() {
			BeforeEach(func() {
				newObj = oldObj.DeepCopy()
			})

			It("should ignore", func() {
				Expect(handler.updateQueue.Len()).To(BeZero())
			})
		})

		When("resource version changed", func() {
			BeforeEach(func() {
				newObj = &v1alpha1.Webhook{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "foo",
						ResourceVersion: "2",
					},
				}
			})

			It("should add to the queue", func() {
				actual, _ := handler.updateQueue.Get()
				Expect(actual).To(Equal(newObj))
			})
		})
	})

	Context("OnDelete", func() {
		expected := &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		}

		BeforeEach(func() {
			handler.deleteQueue = handler.newQueue("delete")
		})

		AfterEach(func() {
			handler.deleteQueue.ShutDown()
		})

		JustBeforeEach(func() {
			handler.OnDelete(expected)
		})

		It("should add to the queue", func() {
			actual, _ := handler.deleteQueue.Get()
			Expect(actual).To(Equal(expected))
		})
	})

	Context("Run", func() {
		var (
			mockCtrl     *gomock.Controller
			eventHandler *MockEventHandler
		)

		webhook := &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test",
				ResourceVersion: "1",
			},
		}

		// Test if TypeMeta is set
		withKind := func(obj *v1alpha1.Webhook) interface{} {
			output := obj.DeepCopy()
			output.SetGroupVersionKind(handler.Kind)
			return output
		}

		testHandler := func(makeCall func() *gomock.Call, trigger func()) {
			waitTimes := func(err error, times int) {
				var wg sync.WaitGroup
				handler.readyCh = make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				By("Start the handler")
				go func() {
					_ = handler.Run(ctx)
				}()
				<-handler.readyCh

				wg.Add(times)
				makeCall().DoAndReturn(func(_ context.Context, _ interface{}) error {
					defer wg.Done()
					return err
				}).Times(times)

				By("Trigger the event")
				trigger()

				By("Wait for calls")
				wg.Wait()

				By("Check if the handler is called")
				mockCtrl.Finish()
			}

			It("should call once", func() {
				waitTimes(nil, 1)
			})

			It("should retry when the handler return errors", func() {
				waitTimes(errors.New("error"), handler.MaxRetry)
			})
		}

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			eventHandler = NewMockEventHandler(mockCtrl)
			handler.recover = GinkgoRecover
			handler.Handler = eventHandler
		})

		When("add", func() {
			testHandler(func() *gomock.Call {
				return eventHandler.EXPECT().OnUpdate(gomock.Any(), withKind(webhook))
			}, func() {
				handler.OnAdd(webhook.DeepCopy())
			})
		})

		When("update", func() {
			testHandler(func() *gomock.Call {
				return eventHandler.EXPECT().OnUpdate(gomock.Any(), withKind(webhook))
			}, func() {
				oldObj := webhook.DeepCopy()
				oldObj.ResourceVersion = "0"
				handler.OnUpdate(oldObj, webhook.DeepCopy())
			})
		})

		When("delete", func() {
			testHandler(func() *gomock.Call {
				return eventHandler.EXPECT().OnDelete(gomock.Any(), withKind(webhook))
			}, func() {
				handler.OnDelete(webhook.DeepCopy())
			})
		})
	})
})
