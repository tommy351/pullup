package k8s

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/internal/testutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Client", func() {
	var client *Client

	getResourceSet := func(name string) (*v1alpha1.ResourceSet, error) {
		return client.Client.PullupV1alpha1().ResourceSets("").Get(name, metav1.GetOptions{})
	}

	Context("GetWebhook", func() {
		var (
			name   string
			actual *v1alpha1.Webhook
			err    error
		)

		webhook := &v1alpha1.Webhook{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		BeforeEach(func() {
			client = &Client{
				Client: fake.NewSimpleClientset(webhook),
			}
		})

		JustBeforeEach(func() {
			actual, err = client.GetWebhook(context.Background(), name)
		})

		When("it exists", func() {
			BeforeEach(func() {
				name = webhook.Name
			})

			It("should return the webhook", func() {
				expected := webhook.DeepCopy()
				expected.APIVersion = v1alpha1.SchemeGroupVersion.String()
				expected.Kind = "Webhook"
				Expect(actual).To(Equal(expected))
			})

			It("should not return the error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("it does not exist", func() {
			BeforeEach(func() {
				name = "not-exist"
			})

			It("should return nil", func() {
				Expect(actual).To(BeNil())
			})

			It("should return the error", func() {
				Expect(IsNotFoundError(err)).To(BeTrue())
			})
		})
	})

	Context("ApplyResourceSet", func() {
		var err error

		input := &v1alpha1.ResourceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.ResourceSetSpec{
				Number: 46,
			},
		}

		getNewResourceSet := func() *v1alpha1.ResourceSet {
			rs, err := getResourceSet(input.Name)
			Expect(err).NotTo(HaveOccurred())
			return rs
		}

		JustBeforeEach(func() {
			err = client.ApplyResourceSet(context.Background(), input)
		})

		When("it exists", func() {
			orig := &v1alpha1.ResourceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					UID:  testutil.NewUID(),
				},
			}

			BeforeEach(func() {
				client = &Client{
					Client: fake.NewSimpleClientset(orig),
				}
			})

			It("should update the resource", func() {
				Expect(getNewResourceSet()).To(Equal(&v1alpha1.ResourceSet{
					ObjectMeta: orig.ObjectMeta,
					Spec:       input.Spec,
				}))
			})

			It("should not return the error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("it does not exist", func() {
			BeforeEach(func() {
				client = &Client{
					Client: fake.NewSimpleClientset(),
				}
			})

			It("should create the resource", func() {
				Expect(getNewResourceSet()).To(Equal(input))
			})

			It("should not return the error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("DeleteResourceSet", func() {
		var (
			name string
			err  error
		)

		rs := &v1alpha1.ResourceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}

		BeforeEach(func() {
			client = &Client{
				Client: fake.NewSimpleClientset(rs),
			}
		})

		JustBeforeEach(func() {
			err = client.DeleteResourceSet(context.Background(), name)
		})

		When("it exists", func() {
			BeforeEach(func() {
				name = rs.Name
			})

			It("should delete the resource", func() {
				_, err := getResourceSet(rs.Name)
				Expect(IsNotFoundError(err)).To(BeTrue())
			})

			It("should not return the error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("it does not exist", func() {
			BeforeEach(func() {
				name = "not-found"
			})

			It("should not delete the resource", func() {
				_, err := getResourceSet(rs.Name)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the error", func() {
				Expect(IsNotFoundError(err)).To(BeTrue())
			})
		})
	})
})
