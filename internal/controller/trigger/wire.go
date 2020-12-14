// +build wireinject

package trigger

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewReconcilerConfig(mgr manager.Manager, logger logr.Logger) ReconcilerConfig {
	wire.Build(
		controller.NewClient,
		controller.NewEventRecorder,
		ReconcilerConfigSet,
	)
	return ReconcilerConfig{}
}
