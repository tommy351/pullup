// +build wireinject

package webhook

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewAlphaReconciler(mgr manager.Manager, logger logr.Logger) *AlphaReconciler {
	wire.Build(
		controller.NewClient,
		controller.NewEventRecorder,
		AlphaReconcilerSet,
	)
	return nil
}

func NewBetaReconcilerConfig(mgr manager.Manager, logger logr.Logger) BetaReconcilerConfig {
	wire.Build(
		controller.NewClient,
		controller.NewEventRecorder,
		NewLogger,
		BetaReconcilerConfigSet,
	)
	return BetaReconcilerConfig{}
}
