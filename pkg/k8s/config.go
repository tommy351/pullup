package k8s

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Namespace string `mapstructure:"namespace"`
	Config    string `mapstructure:"config"`
}

func LoadConfig(config *Config) (*rest.Config, error) {
	if path := config.Config; path != "" {
		return clientcmd.BuildConfigFromFlags("", path)
	}

	return rest.InClusterConfig()
}
