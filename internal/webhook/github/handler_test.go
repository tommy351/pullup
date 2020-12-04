package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// nolint: gochecknoglobals
var (
	fakeRepoOwner         = "foo"
	fakeRepoName          = "bar"
	fakeRepoFullName      = fmt.Sprintf("%s/%s", fakeRepoOwner, fakeRepoName)
	fakePullRequestNumber = 46
)

type pushEventModifier func(event *github.PushEvent)

type pullRequestEventModifier func(event *github.PullRequestEvent)

func fakePushEvent() *github.PushEvent {
	return &github.PushEvent{
		Head: pointer.StringPtr("0ce4cf0450de14c6555c563fa9d36be67e69aa2f"),
		Ref:  pointer.StringPtr("refs/heads/test"),
		Repo: &github.PushEventRepository{
			Name:     &fakeRepoName,
			FullName: &fakeRepoFullName,
			Owner:    &github.User{Login: &fakeRepoOwner},
		},
	}
}

func fakePullRequestEvent() *github.PullRequestEvent {
	return &github.PullRequestEvent{
		Number: &fakePullRequestNumber,
		Action: pointer.StringPtr("opened"),
		Repo: &github.Repository{
			Name:     &fakeRepoName,
			FullName: &fakeRepoFullName,
			Owner:    &github.User{Login: &fakeRepoOwner},
		},
		PullRequest: &github.PullRequest{
			Base: &github.PullRequestBranch{
				Ref: pointer.StringPtr("base"),
				SHA: pointer.StringPtr("b436f6eb3356504235c0c9a8e74605c820d8d9cc"),
			},
			Head: &github.PullRequestBranch{
				Ref: pointer.StringPtr("test"),
				SHA: pointer.StringPtr("0ce4cf0450de14c6555c563fa9d36be67e69aa2f"),
			},
		},
	}
}

var _ = Describe("Handler", func() {
	var (
		handler      *Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
		mgr          *testenv.Manager
		namespaceMap *random.NamespaceMap
	)

	newRequest := func(event string, body interface{}) *http.Request {
		var buf bytes.Buffer
		Expect(json.NewEncoder(&buf).Encode(body)).To(Succeed())
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", event)

		ctx := logr.NewContext(req.Context(), log.Log)

		return req.WithContext(ctx)
	}

	loadTestData := func(name string) []runtime.Object {
		data, err := k8s.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())

		data, err = k8s.MapObjects(data, func(obj runtime.Object) error {
			namespaceMap.SetObject(obj)

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(testenv.CreateObjects(data)).To(Succeed())

		return data
	}

	getChanges := func() []testenv.Change {
		return testenv.GetChanges(handler.Client)
	}

	testGolden := func() {
		It("should match the golden file", func() {
			objects, err := testenv.GetChangedObjects(getChanges())
			Expect(err).NotTo(HaveOccurred())

			objects, err = k8s.MapObjects(objects, func(obj runtime.Object) error {
				namespaceMap.RestoreObject(obj)

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(golden.MatchObject())
		})
	}

	testSuccess := func(name string) {
		var data []runtime.Object

		BeforeEach(func() {
			data = loadTestData(name)
		})

		AfterEach(func() {
			Expect(testenv.DeleteObjects(data)).To(Succeed())
		})

		It("should respond 200", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
		})
	}

	testTriggered := func() {
		It("should change something", func() {
			Expect(getChanges()).NotTo(BeEmpty())
		})
	}

	testSkipped := func() {
		It("should not change anything", func() {
			Expect(getChanges()).To(BeEmpty())
		})
	}

	setPushEvent := func(modifiers ...pushEventModifier) {
		BeforeEach(func() {
			event := fakePushEvent()

			for _, mod := range modifiers {
				mod(event)
			}

			req = newRequest("push", event)
		})
	}

	setPushBranch := func(branch string) pushEventModifier {
		return func(event *github.PushEvent) {
			event.Ref = pointer.StringPtr("refs/heads/" + branch)
		}
	}

	setPushTag := func(tag string) pushEventModifier {
		return func(event *github.PushEvent) {
			event.Ref = pointer.StringPtr("refs/tags/" + tag)
		}
	}

	setPullRequestEvent := func(modifiers ...pullRequestEventModifier) {
		BeforeEach(func() {
			event := fakePullRequestEvent()

			for _, mod := range modifiers {
				mod(event)
			}

			req = newRequest("pull_request", event)
		})
	}

	setPullRequestAction := func(action string) pullRequestEventModifier {
		return func(event *github.PullRequestEvent) {
			event.Action = pointer.StringPtr(action)
		}
	}

	setPullRequestBranch := func(branch string) pullRequestEventModifier {
		return func(event *github.PullRequestEvent) {
			event.PullRequest.Base.Ref = pointer.StringPtr(branch)
		}
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handlerConfig := NewHandlerConfig(Config{}, mgr)
		handler, err = NewHandler(handlerConfig, mgr)
		Expect(err).NotTo(HaveOccurred())

		Expect(mgr.Initialize()).To(Succeed())

		namespaceMap = random.NewNamespaceMap()
	})

	JustBeforeEach(func() {
		recorder = httptest.NewRecorder()
		httputil.NewHandler(handler.Handle)(recorder, req)
	})

	AfterEach(func() {
		mgr.Stop()
	})

	When("payload is invalid", func() {
		BeforeEach(func() {
			req = newRequest("", nil)
		})

		It("should respond 400", func() {
			Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
		})
	})

	Context("v1alpha1.Webhook", func() {
		testApplySuccess := func() {
			When("resource set exists", func() {
				testSuccess("alpha/resource-set-exists")
				testGolden()

				It("should record Updated event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    v1.EventTypeNormal,
						Reason:  hookutil.ReasonUpdated,
						Message: "Updated resource set: foobar-46",
					})).To(BeTrue())
				})
			})

			When("resource set does not exist", func() {
				testSuccess("alpha/resource-set-not-exist")
				testGolden()

				It("should record Created event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    v1.EventTypeNormal,
						Reason:  hookutil.ReasonCreated,
						Message: "Created resource set: foobar-46",
					})).To(BeTrue())
				})
			})
		}

		When("event type is pull request", func() {
			When("no matching webhooks", func() {
				setPullRequestEvent(setPullRequestAction("opened"))

				It("should respond 200", func() {
					Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
				})

				It("should not change anything", func() {
					Expect(getChanges()).To(BeEmpty())
				})
			})

			When("action = opened", func() {
				setPullRequestEvent(setPullRequestAction("opened"))
				testApplySuccess()
			})

			When("action = reopened", func() {
				setPullRequestEvent(setPullRequestAction("reopened"))
				testApplySuccess()
			})

			When("action = synchronize", func() {
				setPullRequestEvent(setPullRequestAction("synchronize"))
				testApplySuccess()
			})

			When("action = closed", func() {
				var data []runtime.Object

				getResourceSetList := func(namespace string, webhookName string, prNumber int) []v1alpha1.ResourceSet {
					list := new(v1alpha1.ResourceSetList)
					err := testenv.GetClient().List(context.Background(), list,
						client.InNamespace(namespaceMap.GetRandom(namespace)),
						client.MatchingLabels(map[string]string{
							k8s.LabelWebhookName:       webhookName,
							k8s.LabelPullRequestNumber: strconv.Itoa(prNumber),
						}))
					Expect(err).NotTo(HaveOccurred())

					return list.Items
				}

				testDeleteSuccess := func() {
					It("should respond 200", func() {
						Expect(recorder.Code).To(Equal(http.StatusOK))
					})

					It("should delete related resource sets", func() {
						Expect(getResourceSetList("test", "foobar", fakePullRequestNumber)).To(BeEmpty())
					})
				}

				setPullRequestEvent(setPullRequestAction("closed"))

				AfterEach(func() {
					Expect(testenv.DeleteObjects(data)).To(Succeed())
				})

				When("resource set exists", func() {
					BeforeEach(func() {
						data = loadTestData("alpha/resource-set-exists")
					})

					testDeleteSuccess()

					It("should record Deleted event", func() {
						Expect(mgr.WaitForEvent(testenv.EventData{
							Type:    v1.EventTypeNormal,
							Reason:  hookutil.ReasonDeleted,
							Message: "Deleted resource sets",
						})).To(BeTrue())
					})
				})

				When("multiple resource sets exist", func() {
					BeforeEach(func() {
						data = loadTestData("alpha/multiple-resource-set")
					})

					testDeleteSuccess()

					It("should not delete resource sets in other namespaces", func() {
						Expect(getResourceSetList("baz", "foobar", 46)).NotTo(BeEmpty())
					})

					It("should not delete resource sets of other pull requests", func() {
						Expect(getResourceSetList("test", "foobar", 47)).NotTo(BeEmpty())
					})

					It("should not delete resource sets of other webhooks", func() {
						Expect(getResourceSetList("test", "baz", 46)).NotTo(BeEmpty())
					})

					It("should record Deleted event", func() {
						Expect(mgr.WaitForEvent(testenv.EventData{
							Type:    v1.EventTypeNormal,
							Reason:  hookutil.ReasonDeleted,
							Message: "Deleted resource sets",
						})).To(BeTrue())
					})
				})

				When("resource set does not exist", func() {
					BeforeEach(func() {
						data = loadTestData("alpha/resource-set-not-exist")
					})

					testDeleteSuccess()

					It("should record Deleted event", func() {
						Expect(mgr.WaitForEvent(testenv.EventData{
							Type:    v1.EventTypeNormal,
							Reason:  hookutil.ReasonDeleted,
							Message: "Deleted resource sets",
						})).To(BeTrue())
					})
				})
			})

			When("resource name is set", func() {
				name := "alpha/resource-name"

				setPullRequestEvent(setPullRequestAction("opened"))
				testSuccess(name)
				testGolden()
			})

			When("branch filter is set", func() {
				When("only include is set", func() {
					name := "alpha/branch-include"

					When("exact match", func() {
						setPullRequestEvent(setPullRequestBranch("foo"))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPullRequestEvent(setPullRequestBranch("bar-5"))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPullRequestEvent(setPullRequestBranch("baz"))
						testSuccess(name)
						testSkipped()
					})
				})

				When("only exclude is set", func() {
					name := "alpha/branch-exclude"

					When("exact match", func() {
						setPullRequestEvent(setPullRequestBranch("foo"))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPullRequestEvent(setPullRequestBranch("bar-5"))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPullRequestEvent(setPullRequestBranch("baz"))
						testSuccess(name)
						testTriggered()
					})
				})

				When("both include and exclude are set", func() {
					name := "alpha/branch-include-exclude"

					When("match include", func() {
						setPullRequestEvent(setPullRequestBranch("ab"))
						testSuccess(name)
						testTriggered()
					})

					When("match both include and exclude", func() {
						setPullRequestEvent(setPullRequestBranch("ac"))
						testSuccess(name)
						testSkipped()
					})

					When("not match include", func() {
						setPullRequestEvent(setPullRequestBranch("a"))
						testSuccess(name)
						testSkipped()
					})
				})
			})
		})
	})

	Context("v1beta1.GitHubWebhook", func() {
		testApplySuccess := func() {
			When("resource exists", func() {
				testSuccess("beta/resource-exists")
				testGolden()

				It("should record Updated event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    v1.EventTypeNormal,
						Reason:  hookutil.ReasonUpdated,
						Message: "Updated resource template: foobar",
					})).To(BeTrue())
				})
			})

			When("resource does not exist", func() {
				testSuccess("beta/resource-not-exist")
				testGolden()

				It("should record Created event", func() {
					Expect(mgr.WaitForEvent(testenv.EventData{
						Type:    v1.EventTypeNormal,
						Reason:  hookutil.ReasonCreated,
						Message: "Created resource template: foobar",
					})).To(BeTrue())
				})
			})
		}

		When("event type = push", func() {
			When("pushing a branch", func() {
				setPushEvent()

				testApplySuccess()

				When("no matching webhooks", func() {
					It("should respond 200", func() {
						Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
					})

					testSkipped()
				})

				When("push event filter is not set", func() {
					testSuccess("beta/without-event-filters")
					testSkipped()
				})

				When("resourceName is not given", func() {
					var data []runtime.Object

					BeforeEach(func() {
						data = loadTestData("beta/without-resource-name")
					})

					AfterEach(func() {
						Expect(testenv.DeleteObjects(data)).To(Succeed())
					})

					It("should respond 400", func() {
						Expect(recorder).To(HaveHTTPStatus(http.StatusBadRequest))
					})
				})

				When("only tag filter is set", func() {
					testSuccess("beta/push-tag-include")
					testSkipped()
				})

				When("branches.include is set", func() {
					name := "beta/push-branch-include"

					When("exact match", func() {
						setPushEvent(setPushBranch("foo"))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPushEvent(setPushBranch("bar-5"))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPushEvent(setPushBranch("abc"))
						testSuccess(name)
						testSkipped()
					})
				})

				When("branches.exclude is set", func() {
					name := "beta/push-branch-exclude"

					When("exact match", func() {
						setPushEvent(setPushBranch("foo"))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPushEvent(setPushBranch("bar-5"))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPushEvent(setPushBranch("abc"))
						testSuccess(name)
						testTriggered()
					})
				})
			})

			When("pushing a tag", func() {
				When("tag filter is not set", func() {
					setPushEvent(setPushTag("foo"))
					testSuccess("beta/resource-not-exist")
					testSkipped()
				})

				When("tags.include is set", func() {
					name := "beta/push-tag-include"

					When("exact match", func() {
						setPushEvent(setPushTag("foo"))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPushEvent(setPushTag("bar-5"))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPushEvent(setPushTag("abc"))
						testSuccess(name)
						testSkipped()
					})
				})

				When("tags.exclude is set", func() {
					name := "beta/push-tag-exclude"

					When("exact match", func() {
						setPushEvent(setPushTag("foo"))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPushEvent(setPushTag("bar-5"))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPushEvent(setPushTag("abc"))
						testSuccess(name)
						testTriggered()
					})
				})
			})
		})

		When("event type = pull_request", func() {
			When("push event filter is not set", func() {
				setPullRequestEvent(setPullRequestAction("opened"))
				testSuccess("beta/without-event-filters")
				testSkipped()
			})

			When("no matching webhooks", func() {
				setPullRequestEvent(setPullRequestAction("opened"))

				It("should respond 200", func() {
					Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
				})

				testSkipped()
			})

			When("action = opened", func() {
				setPullRequestEvent(setPullRequestAction("opened"))
				testApplySuccess()
			})

			When("action = reopened", func() {
				setPullRequestEvent(setPullRequestAction("reopened"))
				testApplySuccess()
			})

			When("action = synchronize", func() {
				setPullRequestEvent(setPullRequestAction("synchronize"))
				testApplySuccess()
			})

			When("action = closed", func() {
				var data []runtime.Object

				setPullRequestEvent(setPullRequestAction("closed"))

				AfterEach(func() {
					Expect(testenv.DeleteObjects(data)).To(Succeed())
				})

				When("resource exists", func() {
					BeforeEach(func() {
						data = loadTestData("beta/resource-exists")
					})

					It("should delete the resource", func() {
						Expect(getChanges()).To(ContainElement(testenv.Change{
							Type: "delete",
							NamespacedName: types.NamespacedName{
								Name:      "foobar",
								Namespace: namespaceMap.GetRandom("test"),
							},
							GroupVersionKind: v1beta1.GroupVersion.WithKind("ResourceTemplate"),
						}))
					})

					It("should record Deleted event", func() {
						Expect(mgr.WaitForEvent(testenv.EventData{
							Type:    v1.EventTypeNormal,
							Reason:  hookutil.ReasonDeleted,
							Message: "Deleted resource template: foobar",
						})).To(BeTrue())
					})
				})

				When("resource does not exist", func() {
					BeforeEach(func() {
						data = loadTestData("beta/resource-not-exist")
					})

					testSkipped()
				})
			})

			When("resourceName is not given", func() {
				setPullRequestEvent(setPullRequestAction("opened"))
				testSuccess("beta/without-resource-name")
				testGolden()
			})

			When("branches.include is set", func() {
				name := "beta/pull-request-branch-include"

				When("exact match", func() {
					setPullRequestEvent(setPullRequestBranch("foo"))
					testSuccess(name)
					testTriggered()
				})

				When("match by regex", func() {
					setPullRequestEvent(setPullRequestBranch("bar-5"))
					testSuccess(name)
					testTriggered()
				})

				When("not match", func() {
					setPullRequestEvent(setPullRequestBranch("abc"))
					testSuccess(name)
					testSkipped()
				})
			})

			When("branches.exclude is set", func() {
				name := "beta/pull-request-branch-exclude"

				When("exact match", func() {
					setPullRequestEvent(setPullRequestBranch("foo"))
					testSuccess(name)
					testSkipped()
				})

				When("match by regex", func() {
					setPullRequestEvent(setPullRequestBranch("bar-5"))
					testSuccess(name)
					testSkipped()
				})

				When("not match", func() {
					setPullRequestEvent(setPullRequestBranch("abc"))
					testSuccess(name)
					testTriggered()
				})
			})
		})
	})
})
