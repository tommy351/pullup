package main

import (
	"fmt"

	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/internal/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager
}

func NewControllerManager(restConf *rest.Config, scheme *runtime.Scheme, conf cmd.Config) (manager.Manager, error) {
	return manager.New(restConf, manager.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: conf.Health.Address,
		MetricsBindAddress:     conf.Metrics.Address,
	})
}

func NewManager(mgr manager.Manager, webhookServer *webhook.Server) (*Manager, error) {
	if err := mgr.Add(webhookServer); err != nil {
		return nil, fmt.Errorf("failed to register the webhook server: %w", err)
	}

	return &Manager{Manager: mgr}, nil
}
