package main

import (
	"fmt"

	"github.com/tommy351/pullup/internal/controller/resourceset"
	"github.com/tommy351/pullup/internal/controller/resourcetemplate"
	"github.com/tommy351/pullup/internal/controller/webhook"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/metrics"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:rbac:groups="",namespace=pullup,resources=configmaps,verbs=get;create;update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch

type Manager struct {
	manager.Manager
}

func NewControllerManager(restConf *rest.Config, scheme *runtime.Scheme, conf k8s.Config) (manager.Manager, error) {
	return manager.New(restConf, manager.Options{
		Scheme:                  scheme,
		LeaderElection:          true,
		LeaderElectionID:        "pullup-controller-lock",
		LeaderElectionNamespace: conf.Namespace,
	})
}

func NewManager(
	mgr manager.Manager,
	rs *resourceset.Reconciler,
	hook *webhook.Reconciler,
	rt *resourcetemplate.Reconciler,
	metricsServer *metrics.Server,
) (*Manager, error) {
	err := builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.Webhook{}).
		Owns(&v1alpha1.ResourceSet{}).
		Complete(hook)
	if err != nil {
		return nil, fmt.Errorf("failed to build the webhook controller: %w", err)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.ResourceSet{}).
		Complete(rs)
	if err != nil {
		return nil, fmt.Errorf("failed to build the resource set controller: %w", err)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1beta1.ResourceTemplate{}).
		Complete(rt)
	if err != nil {
		return nil, fmt.Errorf("failed to build the resource template controller: %w", err)
	}

	if err := mgr.Add(metricsServer); err != nil {
		return nil, fmt.Errorf("failed to register the metrics server: %w", err)
	}

	return &Manager{Manager: mgr}, nil
}
