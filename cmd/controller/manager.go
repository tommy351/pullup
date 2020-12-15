package main

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/internal/controller/resourceset"
	"github.com/tommy351/pullup/internal/controller/resourcetemplate"
	"github.com/tommy351/pullup/internal/controller/trigger"
	"github.com/tommy351/pullup/internal/controller/webhook"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// +kubebuilder:rbac:groups="",namespace=pullup,resources=configmaps,verbs=get;create;update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch

type Manager struct {
	manager.Manager
}

func NewControllerManager(restConf *rest.Config, scheme *runtime.Scheme, conf cmd.Config, logger logr.Logger) (manager.Manager, error) {
	return manager.New(restConf, manager.Options{
		Scheme:                  scheme,
		LeaderElection:          true,
		LeaderElectionID:        "pullup-controller-lock",
		LeaderElectionNamespace: conf.Kubernetes.Namespace,
		HealthProbeBindAddress:  conf.Health.Address,
		MetricsBindAddress:      conf.Metrics.Address,
		Logger:                  logger,
	})
}

func NewManager(
	mgr manager.Manager,
	rs *resourceset.Reconciler,
	hook *webhook.Reconciler,
	rt *resourcetemplate.Reconciler,
	trigger *trigger.Reconciler,
) (*Manager, error) {
	err := builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.Webhook{}).
		Owns(&v1alpha1.ResourceSet{}).
		Complete(hook)
	if err != nil {
		return nil, fmt.Errorf("failed to build Webhook controller: %w", err)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1alpha1.ResourceSet{}).
		Complete(rs)
	if err != nil {
		return nil, fmt.Errorf("failed to build ResourceSet controller: %w", err)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1beta1.ResourceTemplate{}).
		Complete(rt)
	if err != nil {
		return nil, fmt.Errorf("failed to build ResourceTemplate controller: %w", err)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1beta1.Trigger{}).
		Owns(&v1beta1.ResourceTemplate{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to build Trigger controller: %w", err)
	}

	return &Manager{Manager: mgr}, nil
}
