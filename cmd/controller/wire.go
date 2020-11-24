// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/controller/resourceset"
	"github.com/tommy351/pullup/internal/controller/resourcetemplate"
	"github.com/tommy351/pullup/internal/controller/webhook"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/metrics"
)

func InitializeManager(conf cmd.Config) (*Manager, func(), error) {
	wire.Build(
		cmd.ConfigSet,
		log.LoggerSet,
		k8s.Set,
		NewControllerManager,
		controller.NewClient,
		controller.NewEventRecorder,
		resourceset.ReconcilerSet,
		webhook.ReconcilerSet,
		resourcetemplate.ReconcilerSet,
		metrics.ServerSet,
		NewManager,
	)

	return nil, nil, nil
}
