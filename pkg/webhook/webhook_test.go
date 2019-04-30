package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/google/go-github/v24/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("Server.Webhook", func() {
	var (
		client    *testutil.Client
		req       *http.Request
		res       *http.Response
		namespace string
	)

	newRequest := func(body interface{}) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/webhooks/test", &buf)
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	BeforeEach(func() {
		namespace = rand.String(8)
	})

	JustBeforeEach(func() {
		server := &Server{
			Namespace: namespace,
			client:    client,
			logger:    log.NullLogger{},
		}
		router := server.newRouter()
		router.PanicHandler = nil
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		res = recorder.Result()
	})

	When("webhook not found", func() {
		BeforeEach(func() {
			client = testutil.NewClient(env)
			req = newRequest(nil)
		})

		It("should respond 404", func() {
			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should respond error", func() {
			Eventually(gbytes.BufferReader(res.Body)).Should(gbytes.Say("Webhook not found"))
		})
	})

	When("webhook type is unsupported", func() {
		BeforeEach(func() {
			client = testutil.NewClient(env, &v1alpha1.Webhook{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       "Webhook",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
				},
			})
			req = newRequest(nil)
		})

		It("should return 400", func() {
			Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should respond error", func() {
			Eventually(gbytes.BufferReader(res.Body)).Should(gbytes.Say("Unsupported webhook type"))
		})
	})

	When("webhook type is github", func() {
		var (
			webhook     *v1alpha1.Webhook
			resourceSet *v1alpha1.ResourceSet
		)

		BeforeEach(func() {
			webhook = &v1alpha1.Webhook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
					UID:       random.UID(),
				},
				Spec: v1alpha1.WebhookSpec{
					Resources: []json.RawMessage{[]byte("{}")},
					GitHub:    &v1alpha1.GitHubOptions{},
				},
			}
			client = testutil.NewClient(env, webhook)

			resourceSet = &v1alpha1.ResourceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-46",
					Namespace: webhook.Namespace,
					Labels: map[string]string{
						k8s.LabelWebhookName:       webhook.Name,
						k8s.LabelPullRequestNumber: "46",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "pullup.dev/v1alpha1",
							Kind:               "Webhook",
							Name:               webhook.Name,
							UID:                webhook.UID,
							Controller:         pointer.BoolPtr(true),
							BlockOwnerDeletion: pointer.BoolPtr(true),
						},
					},
				},
				Spec: v1alpha1.ResourceSetSpec{},
			}
			Expect(client.Create(context.Background(), resourceSet)).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(client.Delete(context.Background(), webhook)).NotTo(HaveOccurred())
		})

		When("event type is pull request", func() {
			var event *github.PullRequestEvent

			newPullRequestEvent := func(action string) *http.Request {
				num := 46
				event = &github.PullRequestEvent{
					Action: &action,
					Number: &num,
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
				req := newRequest(event)
				req.Header.Set("X-Github-Event", "pull_request")
				return req
			}

			testApply := func() {
				testSuccess := func() {
					It("should respond 204", func() {
						Expect(res.StatusCode).To(Equal(http.StatusNoContent))
					})

					It("should apply the resource set", func() {
						actual := new(v1alpha1.ResourceSet)
						Expect(client.Get(context.TODO(), types.NamespacedName{
							Namespace: resourceSet.Namespace,
							Name:      resourceSet.Name,
						}, actual)).NotTo(HaveOccurred())
						Expect(actual.OwnerReferences).To(Equal(resourceSet.OwnerReferences))
						Expect(actual.Spec).To(Equal(v1alpha1.ResourceSetSpec{
							Resources: webhook.Spec.Resources,
							Number:    event.GetNumber(),
							Base: &v1alpha1.Commit{
								Ref: event.PullRequest.Base.Ref,
								SHA: event.PullRequest.Base.SHA,
							},
							Head: &v1alpha1.Commit{
								Ref: event.PullRequest.Head.Ref,
								SHA: event.PullRequest.Head.SHA,
							},
							Merge: &v1alpha1.Commit{
								SHA: event.PullRequest.MergeCommitSHA,
							},
						}))
					})
				}

				When("resource set exists", func() {
					testSuccess()
				})

				When("resource set does not exist", func() {
					BeforeEach(func() {
						Expect(client.Delete(context.TODO(), resourceSet)).NotTo(HaveOccurred())
					})

					testSuccess()
				})
			}

			When("action is opened", func() {
				BeforeEach(func() {
					req = newPullRequestEvent("opened")
				})

				testApply()
			})

			When("action is reopened", func() {
				BeforeEach(func() {
					req = newPullRequestEvent("reopened")
				})

				testApply()
			})

			When("action is synchronize", func() {
				BeforeEach(func() {
					req = newPullRequestEvent("synchronize")
				})

				testApply()
			})

			When("action is closed", func() {
				BeforeEach(func() {
					req = newPullRequestEvent("closed")
				})

				When("resource set exists", func() {
					It("should respond 204", func() {
						Expect(res.StatusCode).To(Equal(http.StatusNoContent))
					})

					It("should delete the resource set", func() {
						err := client.Get(context.TODO(), types.NamespacedName{
							Namespace: resourceSet.Namespace,
							Name:      resourceSet.Name,
						}, new(v1alpha1.ResourceSet))
						Expect(errors.IsNotFound(err)).To(BeTrue())
					})
				})

				When("resource set not exist", func() {
					BeforeEach(func() {
						Expect(client.Delete(context.TODO(), resourceSet)).NotTo(HaveOccurred())
					})

					It("should respond 204", func() {
						Expect(res.StatusCode).To(Equal(http.StatusNoContent))
					})
				})
			})
		})
	})
})
