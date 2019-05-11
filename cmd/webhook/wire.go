// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/metrics"
	"github.com/tommy351/pullup/internal/webhook"
)

func InitializeManager(conf Config) (*Manager, func(), error) {
	wire.Build(
		NewConfig,
		cmd.ConfigSet,
		log.LoggerSet,
		k8s.Set,
		NewWebhookConfig,
		NewGitHubConfig,
		NewControllerManager,
		webhook.ServerSet,
		metrics.NewServer,
		NewManager,
	)

	return nil, nil, nil
}
