// +build wireinject

package github

import (
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewHandlerConfig(conf Config, mgr manager.Manager) HandlerConfig {
	wire.Build(
		controller.NewClient,
		hookutil.NewEventRecorder,
		hookutil.ResourceTemplateClientSet,
		HandlerConfigSet,
	)
	return HandlerConfig{}
}
