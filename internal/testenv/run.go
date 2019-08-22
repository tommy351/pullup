package testenv

import (
	"path/filepath"
	"testing"

	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration"
)

// Env is a global environment for testing.
// nolint: gochecknoglobals
var Env *Environment

func RunSpecsInEnvironment(t *testing.T, desc string) {
	Env = &Environment{
		ControlPlane: &integration.ControlPlane{
			APIServer: &integration.APIServer{
				Path: AssetBinPath("kube-apiserver"),
			},
			Etcd: &integration.Etcd{
				Path: AssetBinPath("etcd"),
			},
		},
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{
				filepath.Join(ProjectDir(), "deployment", "base", "crds"),
			},
		},
	}

	if err := Env.Start(); err != nil {
		panic(err)
	}

	defer func() {
		_ = Env.Stop()
	}()

	testutil.RunSpecs(t, desc)
}

func GetClient() client.Client {
	return Env.GetClient()
}

func GetScheme() *runtime.Scheme {
	return Env.GetScheme()
}

func NewManager() (*Manager, error) {
	return Env.NewManager()
}
