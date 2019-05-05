// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/pkg/controller/resourceset"
	"github.com/tommy351/pullup/pkg/controller/webhook"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/metrics"
)

func InitializeManager(conf cmd.Config) (*Manager, func(), error) {
	wire.Build(
		cmd.ConfigSet,
		log.LoggerSet,
		k8s.Set,
		NewControllerManager,
		resourceset.NewReconciler,
		webhook.NewReconciler,
		metrics.NewServer,
		NewManager,
	)

	return nil, nil, nil
}
