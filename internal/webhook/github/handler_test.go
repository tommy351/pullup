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
	"github.com/tommy351/pullup/internal/golden"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/random"
	"github.com/tommy351/pullup/internal/testenv"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nolint: gochecknoglobals
var (
	fakeRepoOwner         = "foo"
	fakeRepoName          = "bar"
	fakeRepoFullName      = fmt.Sprintf("%s/%s", fakeRepoOwner, fakeRepoName)
	fakePullRequestNumber = 46
)

func fakePullRequestEvent() *github.PullRequestEvent {
	return &github.PullRequestEvent{
		Number: &fakePullRequestNumber,
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
		Expect(json.NewEncoder(&buf).Encode(body)).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", event)
		return req
	}

	loadTestData := func(name string) []runtime.Object {
		data, err := testutil.LoadObjects(testenv.GetScheme(), fmt.Sprintf("testdata/%s.yml", name))
		Expect(err).NotTo(HaveOccurred())
		data = testutil.MapObjects(data, namespaceMap.SetObject)
		Expect(testenv.CreateObjects(data)).To(Succeed())
		mgr.WaitForSync()
		return data
	}

	getChanges := func() []testenv.Change {
		return testenv.GetChanges(handler.client)
	}

	testGolden := func(name string) {
		It("should match the golden file", func() {
			objects, err := testenv.GetChangedObjects(getChanges())
			Expect(err).NotTo(HaveOccurred())
			objects = testutil.MapObjects(objects, namespaceMap.RestoreObject)
			Expect(objects).To(golden.MatchObject(fmt.Sprintf("testdata/%s.golden", name)))
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

		It("should respond 204", func() {
			Expect(recorder.Code).To(Equal(http.StatusNoContent))
		})
	}

	testApplySuccess := func() {
		When("resource set exists", func() {
			testSuccess("resource-set-exists")
			testGolden("apply-success")
		})

		When("resource set does not exist", func() {
			testSuccess("resource-set-not-exist")
			testGolden("apply-success")
		})
	}

	BeforeEach(func() {
		var err error
		mgr, err = testenv.NewManager()
		Expect(err).NotTo(HaveOccurred())

		handler, err = NewHandler(Config{}, mgr)
		Expect(err).NotTo(HaveOccurred())

		Expect(mgr.Initialize()).To(Succeed())
		mgr.WaitForSync()

		namespaceMap = random.NewNamespaceMap()
	})

	AfterEach(func() {
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
		setRequest := func(action string) {
			BeforeEach(func() {
				event := fakePullRequestEvent()
				event.Action = pointer.StringPtr(action)
				req = newRequest("pull_request", event)
			})
		}

		When("no matching webhooks", func() {
			setRequest("opened")

			It("should respond 204", func() {
				Expect(recorder.Code).To(Equal(http.StatusNoContent))
			})

			It("should not change anything", func() {
				Expect(getChanges()).To(BeEmpty())
			})
		})

		When("action = opened", func() {
			setRequest("opened")
			testApplySuccess()
		})

		When("action = reopened", func() {
			setRequest("reopened")
			testApplySuccess()
		})

		When("action = synchronize", func() {
			setRequest("synchronize")
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
				It("should respond 204", func() {
					Expect(recorder.Code).To(Equal(http.StatusNoContent))
				})

				It("should delete related resource sets", func() {
					Expect(getResourceSetList("test", "foobar", fakePullRequestNumber)).To(BeEmpty())
				})
			}

			setRequest("closed")

			AfterEach(func() {
				Expect(testenv.DeleteObjects(data)).To(Succeed())
			})

			When("resource set exists", func() {
				BeforeEach(func() {
					data = loadTestData("resource-set-exists")
				})

				testDeleteSuccess()
			})

			When("multiple resource sets exist", func() {
				BeforeEach(func() {
					data = loadTestData("multiple-resource-set")
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
			})

			When("resource set does not exist", func() {
				BeforeEach(func() {
					data = loadTestData("resource-set-not-exist")
				})

				testDeleteSuccess()
			})
		})

		When("resource name is set", func() {
			name := "resource-name"

			setRequest("opened")
			testSuccess(name)
			testGolden(name)
		})

		When("branch filter is set", func() {
			setBranch := func(branch string) {
				BeforeEach(func() {
					event := fakePullRequestEvent()
					event.Action = pointer.StringPtr("opened")
					event.PullRequest.Base.Ref = pointer.StringPtr(branch)
					req = newRequest("pull_request", event)
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

			When("only include is set", func() {
				name := "branch-include"

				When("exact match", func() {
					setBranch("foo")
					testSuccess(name)
					testTriggered()
				})

				When("match by regex", func() {
					setBranch("bar-5")
					testSuccess(name)
					testTriggered()
				})

				When("not match", func() {
					setBranch("baz")
					testSuccess(name)
					testSkipped()
				})
			})

			When("only exclude is set", func() {
				name := "branch-exclude"

				When("exact match", func() {
					setBranch("foo")
					testSuccess(name)
					testSkipped()
				})

				When("match by regex", func() {
					setBranch("bar-5")
					testSuccess(name)
					testSkipped()
				})

				When("not match", func() {
					setBranch("baz")
					testSuccess(name)
					testTriggered()
				})
			})

			When("both include and exclude are set", func() {
				name := "branch-include-exclude"

				When("match include", func() {
					setBranch("ab")
					testSuccess(name)
					testTriggered()
				})

				When("match both include and exclude", func() {
					setBranch("ac")
					testSuccess(name)
					testSkipped()
				})

				When("not match include", func() {
					setBranch("a")
					testSuccess(name)
					testSkipped()
				})
			})
		})
	})
})
