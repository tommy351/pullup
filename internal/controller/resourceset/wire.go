// +build wireinject

package resourceset

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewReconciler(mgr manager.Manager, logger logr.Logger) *Reconciler {
	wire.Build(
		controller.NewClient,
		controller.NewEventRecorder,
		ReconcilerSet,
	)
	return nil
}
