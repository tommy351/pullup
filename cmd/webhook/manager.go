package main

import (
	"github.com/tommy351/pullup/internal/metrics"
	"github.com/tommy351/pullup/internal/webhook"
	"github.com/tommy351/pullup/internal/webhook/github"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager
}

func NewWebhookConfig(conf Config) webhook.Config {
	return conf.Webhook
}

func NewGitHubConfig(conf Config) github.Config {
	return conf.GitHub
}

func NewControllerManager(restConf *rest.Config, scheme *runtime.Scheme) (manager.Manager, error) {
	return manager.New(restConf, manager.Options{
		Scheme: scheme,
	})
}

func NewManager(mgr manager.Manager, webhookServer *webhook.Server, metricsServer *metrics.Server) (*Manager, error) {
	if err := mgr.Add(webhookServer); err != nil {
		return nil, err
	}

	if err := mgr.Add(metricsServer); err != nil {
		return nil, err
	}

	return &Manager{Manager: mgr}, nil
}
