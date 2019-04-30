// +build integration

package testutil

import (
	"path/filepath"

	"github.com/tommy351/pullup/internal/testenv"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration"
)

func NewEnvironment() testenv.Interface {
	return &testenv.Integration{
		Scheme: Scheme,
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
				filepath.Join(ProjectDir(), "deployment", "crds"),
			},
		},
	}
}
