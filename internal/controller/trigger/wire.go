// +build wireinject

package trigger

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewReconcilerConfig(mgr manager.Manager) ReconcilerConfig {
	wire.Build(
		controller.NewClient,
		controller.NewEventRecorder,
		ReconcilerConfigSet,
	)
	return ReconcilerConfig{}
}
