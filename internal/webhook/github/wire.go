// +build wireinject

package github

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewHandler(conf Config, mgr manager.Manager) *Handler {
	wire.Build(
		controller.NewClient,
		hookutil.NewEventRecorder,
		hookutil.NewFieldIndexer,
		hookutil.ResourceTemplateClientSet,
		HandlerSet,
	)
	return nil
}
