package k8s

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Load auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Set provides config and scheme.
// nolint: gochecknoglobals
var Set = wire.NewSet(LoadConfig, NewScheme)

type Config struct {
	Namespace string `mapstructure:"namespace"`
	Config    string `mapstructure:"config"`
}

func LoadConfig(config Config) (*rest.Config, error) {
	if path := config.Config; path != "" {
		return clientcmd.BuildConfigFromFlags("", path)
	}

	return rest.InClusterConfig()
}

func NewScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	sb := runtime.NewSchemeBuilder(scheme.AddToScheme, v1alpha1.AddToScheme)

	if err := sb.AddToScheme(s); err != nil {
		return nil, err
	}

	return s, nil
}
