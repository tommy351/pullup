package builder

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

type Builder struct {
	b      *builder.Builder
	logger logr.Logger
}

func New(mgr manager.Manager) *Builder {
	return &Builder{
		b: builder.ControllerManagedBy(mgr),
	}
}

func (b *Builder) For(apiType runtime.Object) *Builder {
	b.b.For(apiType)
	return b
}

func (b *Builder) Owns(apiType runtime.Object) *Builder {
	b.b.Owns(apiType)
	return b
}

func (b *Builder) WithLogger(logger logr.Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) Complete(r reconcile.Reconciler) error {
	if b.logger != nil {
		if _, err := inject.LoggerInto(b.logger, r); err != nil {
			return err
		}
	}

	return b.b.Complete(r)
}
