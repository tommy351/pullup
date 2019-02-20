package kubernetes

import (
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadConfig() (*rest.Config, error) {
	conf, err := rest.InClusterConfig()

	if err == nil {
		return conf, nil
	} else if err != rest.ErrNotInCluster {
		return nil, err
	}

	home, err := homedir.Dir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", path)
}

func GetVersionedConfig(input *rest.Config, apiVersion string) *rest.Config {
	var gv schema.GroupVersion
	conf := rest.CopyConfig(input)
	parts := strings.SplitN(apiVersion, "/", 2)

	if len(parts) == 2 {
		gv.Group = parts[0]
		gv.Version = parts[1]
	} else {
		gv.Version = parts[0]
	}

	conf.GroupVersion = &gv

	if gv.Group == "" {
		conf.APIPath = "/api"
	} else {
		conf.APIPath = "/apis"
	}

	return conf
}
