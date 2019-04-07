package fake

import (
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned/fake"
	"github.com/tommy351/pullup/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewClient(objects ...runtime.Object) *k8s.Client {
	scheme := runtime.NewScheme()

	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme); err != nil {
		panic(err)
	}

	return &k8s.Client{
		Namespace: "default",
		Dynamic:   dynamicfake.NewSimpleDynamicClient(scheme, objects...),
		Client:    fake.NewSimpleClientset(objects...),
	}
}
