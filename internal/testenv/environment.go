package testenv

import (
	"github.com/tommy351/pullup/internal/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/testing_frameworks/integration"
)

type Environment struct {
	ControlPlane      *integration.ControlPlane
	CRDInstallOptions envtest.CRDInstallOptions

	config *rest.Config
	client client.Client
	scheme *runtime.Scheme
}

func (e *Environment) Start() (err error) {
	if e.scheme, err = k8s.NewScheme(); err != nil {
		return
	}

	if err = e.ControlPlane.Start(); err != nil {
		return
	}

	e.config = &rest.Config{
		Host: e.ControlPlane.APIURL().Host,
	}

	if _, err = envtest.InstallCRDs(e.config, e.CRDInstallOptions); err != nil {
		return
	}

	if e.client, err = client.New(e.config, client.Options{Scheme: e.scheme}); err != nil {
		return
	}

	return
}

func (e *Environment) Stop() error {
	return e.ControlPlane.Stop()
}

func (e *Environment) GetConfig() *rest.Config {
	return e.config
}

func (e *Environment) GetClient() client.Client {
	return e.client
}

func (e *Environment) GetScheme() *runtime.Scheme {
	return e.scheme
}

func (e *Environment) NewManager() (*Manager, error) {
	m, err := manager.New(e.config, manager.Options{
		Scheme: e.scheme,
	})

	if err != nil {
		return nil, err
	}

	return &Manager{Manager: m}, nil
}
