package k8s

import (
	"os"
	"path/filepath"

	"github.com/ansel1/merry"
	homedir "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadConfig() (*rest.Config, error) {
	path := getKubeConfigPath()
	conf, err := clientcmd.BuildConfigFromFlags("", path)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return conf, nil
}

func getKubeConfigPath() string {
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}

	home, err := homedir.Dir()

	if err != nil {
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}
