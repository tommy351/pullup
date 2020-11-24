package resourcetemplate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcilerSet provides a reconciler.
// nolint: gochecknoglobals
var ReconcilerSet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Reconciler), "*"),
)

type Logger logr.Logger

func NewLogger(logger logr.Logger) Logger {
	return logger.WithName("controller").WithName("resourcetemplate")
}

type Reconciler struct {
	Client client.Client
	Logger Logger
}

func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	rt := new(v1beta1.ResourceTemplate)
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, rt); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get resource template: %w", err)
	}

	return reconcile.Result{}, nil
}
