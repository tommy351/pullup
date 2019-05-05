package main

import (
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/controller/resourceset"
	"github.com/tommy351/pullup/pkg/controller/webhook"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/metrics"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager
}

func NewControllerManager(restConf *rest.Config, scheme *runtime.Scheme, conf k8s.Config) (manager.Manager, error) {
	return manager.New(restConf, manager.Options{
		Scheme:                  scheme,
		LeaderElection:          true,
		LeaderElectionNamespace: conf.Namespace,
	})
}

func NewManager(mgr manager.Manager, rs *resourceset.Reconciler, hook *webhook.Reconciler, metricsServer *metrics.Server) (*Manager, error) {
	err := builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.Webhook{}).
		Owns(&v1alpha1.ResourceSet{}).
		Complete(hook)

	if err != nil {
		return nil, err
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.ResourceSet{}).
		Complete(rs)

	if err != nil {
		return nil, err
	}

	if err := mgr.Add(metricsServer); err != nil {
		return nil, err
	}

	return &Manager{Manager: mgr}, nil
}
