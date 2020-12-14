// +build wireinject

package hookutil

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewTriggerHandler(mgr manager.Manager) *TriggerHandler {
	wire.Build(
		controller.NewClient,
		NewEventRecorder,
		TriggerHandlerSet,
	)
	return nil
}
