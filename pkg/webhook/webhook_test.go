package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/google/go-github/v24/github"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/k8s/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("Server.Webhook", func() {
	var (
		client *k8s.Client
		req    *http.Request
		res    *http.Response
	)

	newRequest := func(body interface{}) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/webhooks/test", &buf)
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	JustBeforeEach(func() {
		server := &Server{
			Client: client,
		}
		router := server.newRouter()
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		res = recorder.Result()
	})

	When("webhook not found", func() {
		BeforeEach(func() {
			client = fake.NewClient()
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
			client = fake.NewClient(&v1alpha1.Webhook{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       "Webhook",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
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
		webhook := &v1alpha1.Webhook{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       "Webhook",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				UID:       types.UID(uuid.Must(uuid.NewRandom()).String()),
			},
			Spec: v1alpha1.WebhookSpec{
				Resources: []json.RawMessage{[]byte("{}")},
				GitHub:    &v1alpha1.GitHubOptions{},
			},
		}

		resourceSet := &v1alpha1.ResourceSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       "ResourceSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-46",
				Namespace: webhook.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         webhook.APIVersion,
						Kind:               webhook.Kind,
						Name:               webhook.Name,
						UID:                webhook.UID,
						Controller:         pointer.BoolPtr(true),
						BlockOwnerDeletion: pointer.BoolPtr(true),
					},
				},
			},
			Spec: v1alpha1.ResourceSetSpec{},
		}

		BeforeEach(func() {
			client = fake.NewClient(webhook, resourceSet)
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
							SHA: pointer.StringPtr(testutil.RandomSHA1()),
						},
						Head: &github.PullRequestBranch{
							Ref: pointer.StringPtr("test"),
							SHA: pointer.StringPtr(testutil.RandomSHA1()),
						},
						MergeCommitSHA: pointer.StringPtr(testutil.RandomSHA1()),
					},
				}
				req := newRequest(event)
				req.Header.Set("X-Github-Event", "pull_request")
				return req
			}

			testApply := func() {
				It("should respond 204", func() {
					Expect(res.StatusCode).To(Equal(http.StatusNoContent))
				})

				It("should apply the resource set", func() {
					rs, err := client.Client.PullupV1alpha1().ResourceSets(resourceSet.Namespace).Get(resourceSet.Name, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())
					Expect(rs).To(Equal(&v1alpha1.ResourceSet{
						TypeMeta:   resourceSet.TypeMeta,
						ObjectMeta: resourceSet.ObjectMeta,
						Spec: v1alpha1.ResourceSetSpec{
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
						},
					}))
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
						_, err := client.Client.PullupV1alpha1().ResourceSets(resourceSet.Namespace).Get(resourceSet.Name, metav1.GetOptions{})
						Expect(k8s.IsNotFoundError(err)).To(BeTrue())
					})
				})

				When("resource set not exist", func() {
					BeforeEach(func() {
						Expect(client.Client.PullupV1alpha1().ResourceSets(resourceSet.Namespace).Delete(resourceSet.Name, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
					})

					It("should respond 204", func() {
						Expect(res.StatusCode).To(Equal(http.StatusNoContent))
					})
				})
			})
		})
	})
})
