// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/controller/resourceset"
	"github.com/tommy351/pullup/internal/controller/resourcetemplate"
	"github.com/tommy351/pullup/internal/controller/trigger"
	"github.com/tommy351/pullup/internal/controller/webhook"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
)

func InitializeManager(conf cmd.Config) (*Manager, func(), error) {
	wire.Build(
		cmd.ConfigSet,
		log.LoggerSet,
		k8s.Set,
		NewControllerManager,
		controller.NewClient,
		controller.NewEventRecorder,
		controller.NewAPIReader,
		resourceset.ReconcilerSet,
		webhook.ReconcilerSet,
		trigger.ReconcilerSet,
		resourcetemplate.ReconcilerSet,
		NewManager,
	)

	return nil, nil, nil
}
