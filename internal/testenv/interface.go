package testenv

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Interface interface {
	NewClient(objects ...runtime.Object) (client.Client, error)
	Start() error
	Stop() error
}
