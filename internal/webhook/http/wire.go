// +build wireinject

package http

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewHandler(mgr manager.Manager) *Handler {
	wire.Build(
		controller.NewClient,
		hookutil.NewEventRecorder,
		hookutil.TriggerHandlerSet,
		HandlerSet,
	)
	return nil
}
