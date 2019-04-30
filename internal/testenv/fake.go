package testenv

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Fake struct {
	Scheme *runtime.Scheme
}

func (f *Fake) NewClient(objects ...runtime.Object) (client.Client, error) {
	return fake.NewFakeClientWithScheme(f.Scheme, objects...), nil
}

func (*Fake) Start() error {
	return nil
}

func (*Fake) Stop() error {
	return nil
}
