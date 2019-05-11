package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/google/go-github/v25/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("Handler", func() {
	var (
		handler   *Handler
		req       *http.Request
		recorder  *httptest.ResponseRecorder
		namespace string
		webhook   *v1alpha1.Webhook
		mgr       *testenv.Manager
	)

	newRequest := func(event string, body interface{}) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", event)
		return req
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handler, err = NewHandler(Config{}, mgr)
		Expect(err).NotTo(HaveOccurred())

		Expect(mgr.Initialize()).To(Succeed())

		namespace = random.Namespace()
		webhook = &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "foobar",
			},
			Spec: v1alpha1.WebhookSpec{
				Resources: []json.RawMessage{[]byte("{}")},
				Repositories: []v1alpha1.WebhookRepository{
					{
						Type: "github",
						Name: "foo/bar",
					},
				},
			},
		}
		Expect(testenv.GetClient().Create(context.Background(), webhook)).To(Succeed())
		mgr.WaitForSync()
	})

	AfterEach(func() {
		Expect(testenv.GetClient().Delete(context.Background(), webhook)).To(Succeed())
		mgr.Stop()
	})

	JustBeforeEach(func() {
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
	})

	When("payload is invalid", func() {
		BeforeEach(func() {
			req = newRequest("", nil)
		})

		It("should respond 400", func() {
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	When("event type is pull request", func() {
		var event *github.PullRequestEvent

		newPullRequestEvent := func(action string) *github.PullRequestEvent {
			num := 46
			return &github.PullRequestEvent{
				Action: pointer.StringPtr(action),
				Number: &num,
				Repo: &github.Repository{
					FullName: pointer.StringPtr("foo/bar"),
				},
				PullRequest: &github.PullRequest{
					Base: &github.PullRequestBranch{
						Ref: pointer.StringPtr("master"),
						SHA: pointer.StringPtr(random.SHA1()),
					},
					Head: &github.PullRequestBranch{
						Ref: pointer.StringPtr("test"),
						SHA: pointer.StringPtr(random.SHA1()),
					},
					MergeCommitSHA: pointer.StringPtr(random.SHA1()),
				},
			}
		}

		testApply := func() {
			testSuccess := func() {
				It("should respond 204", func() {
					Expect(recorder.Code).To(Equal(http.StatusNoContent))
				})

				It("should apply the resource set", func() {
					obj := new(v1alpha1.ResourceSet)
					err := testenv.GetClient().Get(context.Background(), types.NamespacedName{
						Namespace: namespace,
						Name:      fmt.Sprintf("%s-%d", webhook.Name, event.GetNumber()),
					}, obj)
					Expect(err).NotTo(HaveOccurred())
					Expect(obj.Labels).To(Equal(map[string]string{
						k8s.LabelWebhookName:       webhook.Name,
						k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
					}))
					Expect(obj.OwnerReferences).To(ConsistOf([]metav1.OwnerReference{
						{
							APIVersion:         v1alpha1.SchemeGroupVersion.String(),
							Kind:               "Webhook",
							Name:               webhook.Name,
							UID:                webhook.UID,
							Controller:         pointer.BoolPtr(true),
							BlockOwnerDeletion: pointer.BoolPtr(true),
						},
					}))
					Expect(obj.Spec).To(Equal(v1alpha1.ResourceSetSpec{
						Resources: webhook.Spec.Resources,
						Number:    event.GetNumber(),
						Base: v1alpha1.Commit{
							Ref: event.PullRequest.Base.GetRef(),
							SHA: event.PullRequest.Base.GetSHA(),
						},
						Head: v1alpha1.Commit{
							Ref: event.PullRequest.Head.GetRef(),
							SHA: event.PullRequest.Head.GetSHA(),
						},
					}))
				})
			}

			When("resource set exists", func() {
				var rs *v1alpha1.ResourceSet

				BeforeEach(func() {
					rs = &v1alpha1.ResourceSet{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: namespace,
							Name:      fmt.Sprintf("%s-%d", webhook.Name, event.GetNumber()),
							Labels: map[string]string{
								k8s.LabelWebhookName:       webhook.Name,
								k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
							},
							OwnerReferences: []metav1.OwnerReference{
								{
									APIVersion:         v1alpha1.SchemeGroupVersion.String(),
									Kind:               "Webhook",
									Name:               webhook.Name,
									UID:                webhook.UID,
									Controller:         pointer.BoolPtr(true),
									BlockOwnerDeletion: pointer.BoolPtr(true),
								},
							},
						},
					}
					Expect(testenv.GetClient().Create(context.Background(), rs)).To(Succeed())
				})

				testSuccess()
			})

			When("resource set does not exist", func() {
				testSuccess()
			})
		}

		When("no matching webhooks", func() {
			BeforeEach(func() {
				event = newPullRequestEvent("opened")
				event.Repo.FullName = pointer.StringPtr("bar/foo")
				req = newRequest("pull_request", event)
			})

			It("should respond 204", func() {
				Expect(recorder.Code).To(Equal(http.StatusNoContent))
			})

			It("should do nothing", func() {
				Expect(testenv.GetChanges(handler.client)).To(BeEmpty())
			})
		})

		When("action = opened", func() {
			BeforeEach(func() {
				event = newPullRequestEvent("opened")
				req = newRequest("pull_request", event)
			})

			testApply()
		})

		When("action = reopened", func() {
			BeforeEach(func() {
				event = newPullRequestEvent("reopened")
				req = newRequest("pull_request", event)
			})

			testApply()
		})

		When("action = synchronize", func() {
			BeforeEach(func() {
				event = newPullRequestEvent("synchronize")
				req = newRequest("pull_request", event)
			})

			testApply()
		})

		When("action = closed", func() {
			BeforeEach(func() {
				event = newPullRequestEvent("closed")
				req = newRequest("pull_request", event)
			})

			When("resource set exists", func() {
				var rs *v1alpha1.ResourceSet

				BeforeEach(func() {
					rs = &v1alpha1.ResourceSet{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: namespace,
							Name:      fmt.Sprintf("%s-%d", webhook.Name, event.GetNumber()),
						},
					}
					Expect(testenv.GetClient().Create(context.Background(), rs)).To(Succeed())
				})

				It("should respond 204", func() {
					Expect(recorder.Code).To(Equal(http.StatusNoContent))
				})

				It("should delete the resource set", func() {
					err := testenv.GetClient().Get(context.Background(), types.NamespacedName{
						Namespace: rs.Namespace,
						Name:      rs.Name,
					}, new(v1alpha1.ResourceSet))
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			When("resource set does not exist", func() {
				It("should respond 204", func() {
					Expect(recorder.Code).To(Equal(http.StatusNoContent))
				})
			})
		})
	})
})
