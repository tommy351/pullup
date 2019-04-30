package testutil

import (
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// Scheme stores necessary schemes for testing.
// nolint: gochecknoglobals
var Scheme = runtime.NewScheme()

// nolint: gochecknoinits
func init() {
	sb := runtime.NewSchemeBuilder(scheme.AddToScheme, v1alpha1.AddToScheme)

	if err := sb.AddToScheme(Scheme); err != nil {
		panic(err)
	}
}
