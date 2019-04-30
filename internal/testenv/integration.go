package testenv

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration"
)

type Integration struct {
	*integration.ControlPlane

	Scheme            *runtime.Scheme
	CRDInstallOptions envtest.CRDInstallOptions
}

func (i *Integration) restConfig() *rest.Config {
	return &rest.Config{
		Host: i.ControlPlane.APIURL().Host,
	}
}

func (i *Integration) Start() error {
	if err := i.ControlPlane.Start(); err != nil {
		return err
	}

	_, err := envtest.InstallCRDs(i.restConfig(), i.CRDInstallOptions)
	return err
}

func (i *Integration) NewClient(objects ...runtime.Object) (client.Client, error) {
	conf := i.restConfig()
	mapper, err := apiutil.NewDiscoveryRESTMapper(conf)

	if err != nil {
		return nil, err
	}

	c, err := client.New(conf, client.Options{
		Scheme: i.Scheme,
		Mapper: mapper,
	})

	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		if err := c.Create(context.Background(), obj); err != nil {
			return nil, err
		}
	}

	return c, nil
}
