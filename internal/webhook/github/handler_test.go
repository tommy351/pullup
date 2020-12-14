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
	"github.com/tommy351/pullup/internal/fakegithub"
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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

		data, err = k8s.MapObjects(data, namespaceMap.SetObject)
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

			objects, err = k8s.MapObjects(objects, namespaceMap.RestoreObject)
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

	setPullRequestEvent := func(event *github.PullRequestEvent) {
		BeforeEach(func() {
			req = newRequest("pull_request", event)
		})
	}

	setPushEvent := func(event *github.PushEvent) {
		BeforeEach(func() {
			req = newRequest("push", event)
		})
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
		hookutil.NewHandler(handler.Handle).ServeHTTP(recorder, req)
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
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))

				It("should respond 200", func() {
					Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
				})

				It("should not change anything", func() {
					Expect(getChanges()).To(BeEmpty())
				})
			})

			When("action = opened", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))
				testApplySuccess()
			})

			When("action = reopened", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("reopened")))
				testApplySuccess()
			})

			When("action = synchronize", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("synchronize")))
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
						Expect(getResourceSetList("test", "foobar", 46)).To(BeEmpty())
					})
				}

				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("closed")))

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

				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))
				testSuccess(name)
				testGolden()
			})

			When("branch filter is set", func() {
				When("only include is set", func() {
					name := "alpha/branch-include"

					When("exact match", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("foo")))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("bar-5")))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("baz")))
						testSuccess(name)
						testSkipped()
					})
				})

				When("only exclude is set", func() {
					name := "alpha/branch-exclude"

					When("exact match", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("foo")))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("bar-5")))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("baz")))
						testSuccess(name)
						testTriggered()
					})
				})

				When("both include and exclude are set", func() {
					name := "alpha/branch-include-exclude"

					When("match include", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("ab")))
						testSuccess(name)
						testTriggered()
					})

					When("match both include and exclude", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("ac")))
						testSuccess(name)
						testSkipped()
					})

					When("not match include", func() {
						setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("a")))
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
				setPushEvent(fakegithub.NewPushEvent())

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

				When("only tag filter is set", func() {
					testSuccess("beta/push-tag-include")
					testSkipped()
				})

				When("branches.include is set", func() {
					name := "beta/push-branch-include"

					When("exact match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("foo")))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("bar-5")))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("abc")))
						testSuccess(name)
						testSkipped()
					})
				})

				When("branches.exclude is set", func() {
					name := "beta/push-branch-exclude"

					When("exact match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("foo")))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("bar-5")))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventBranch("abc")))
						testSuccess(name)
						testTriggered()
					})
				})
			})

			When("pushing a tag", func() {
				When("tag filter is not set", func() {
					setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("foo")))
					testSuccess("beta/resource-not-exist")
					testSkipped()
				})

				When("tags.include is set", func() {
					name := "beta/push-tag-include"

					When("exact match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("foo")))
						testSuccess(name)
						testTriggered()
					})

					When("match by regex", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("bar-5")))
						testSuccess(name)
						testTriggered()
					})

					When("not match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("abc")))
						testSuccess(name)
						testSkipped()
					})
				})

				When("tags.exclude is set", func() {
					name := "beta/push-tag-exclude"

					When("exact match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("foo")))
						testSuccess(name)
						testSkipped()
					})

					When("match by regex", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("bar-5")))
						testSuccess(name)
						testSkipped()
					})

					When("not match", func() {
						setPushEvent(fakegithub.NewPushEvent(fakegithub.SetPushEventTag("abc")))
						testSuccess(name)
						testTriggered()
					})
				})
			})

			When("action is set", func() {
				setPushEvent(fakegithub.NewPushEvent())
				testSuccess("beta/action")

				It("should not have any changes", func() {
					Expect(getChanges()).To(BeEmpty())
				})
			})
		})

		When("event type = pull_request", func() {
			When("push event filter is not set", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))
				testSuccess("beta/without-event-filters")
				testSkipped()
			})

			When("no matching webhooks", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))

				It("should respond 200", func() {
					Expect(recorder).To(HaveHTTPStatus(http.StatusOK))
				})

				testSkipped()
			})

			When("action = opened", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))
				testApplySuccess()
			})

			When("action = reopened", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("reopened")))
				testApplySuccess()
			})

			When("action = synchronize", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("synchronize")))
				testApplySuccess()
			})

			When("action = closed", func() {
				var data []runtime.Object

				setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("closed")))

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

			When("branches.include is set", func() {
				name := "beta/pull-request-branch-include"

				When("exact match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("foo")))
					testSuccess(name)
					testTriggered()
				})

				When("match by regex", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("bar-5")))
					testSuccess(name)
					testTriggered()
				})

				When("not match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("abc")))
					testSuccess(name)
					testSkipped()
				})
			})

			When("branches.exclude is set", func() {
				name := "beta/pull-request-branch-exclude"

				When("exact match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("foo")))
					testSuccess(name)
					testSkipped()
				})

				When("match by regex", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("bar-5")))
					testSuccess(name)
					testSkipped()
				})

				When("not match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventBranch("abc")))
					testSuccess(name)
					testTriggered()
				})
			})

			When("labels.include is set", func() {
				name := "beta/pull-request-label-include"

				When("exact match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"foo"})))
					testSuccess(name)
					testTriggered()
				})

				When("match by regex", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"bar-5"})))
					testSuccess(name)
					testTriggered()
				})

				When("not match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"abc"})))
					testSuccess(name)
					testSkipped()
				})

				When("label is present in pull request event", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(
						fakegithub.SetPullRequestEventTriggeredLabel("foo"),
					))
					testSuccess(name)
					testTriggered()
				})
			})

			When("labels.exclude is set", func() {
				name := "beta/pull-request-label-exclude"

				When("exact match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"foo"})))
					testSuccess(name)
					testSkipped()
				})

				When("match by regex", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"bar-5"})))
					testSuccess(name)
					testSkipped()
				})

				When("not match", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventLabels([]string{"abc"})))
					testSuccess(name)
					testTriggered()
				})
			})

			When("types is set", func() {
				name := "beta/pull-request-type"

				When("action = labeled", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("labeled")))
					testSuccess(name)
					testTriggered()
				})

				When("action = unlabeled", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("unlabeled")))
					testSuccess(name)
					testSkipped()
				})

				When("action = opened", func() {
					setPullRequestEvent(fakegithub.NewPullRequestEvent(fakegithub.SetPullRequestEventAction("opened")))
					testSuccess(name)
					testSkipped()
				})
			})

			When("action is set", func() {
				setPullRequestEvent(fakegithub.NewPullRequestEvent())
				testSuccess("beta/action")

				It("should not have any changes", func() {
					Expect(getChanges()).To(BeEmpty())
				})
			})
		})
	})
})
