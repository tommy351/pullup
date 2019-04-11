package manager

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/client/informers/externalversions"
	"github.com/tommy351/pullup/pkg/group"
	"k8s.io/client-go/dynamic"
)

type Manager struct {
	Namespace string
	Informer  externalversions.SharedInformerFactory
	Dynamic   dynamic.Interface
}

func (m *Manager) Run(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	g := group.NewGroup(ctx)

	rsHandler := &Handler{
		Kind:     v1alpha1.Kind("ResourceSet"),
		MaxRetry: 5,
		Handler: &ResourceSetEventHandler{
			Namespace: m.Namespace,
			Dynamic:   m.Dynamic,
		},
	}

	m.Informer.Pullup().V1alpha1().ResourceSets().Informer().AddEventHandler(rsHandler)

	g.Go(rsHandler.Run)

	g.Go(func(ctx context.Context) error {
		logger.Debug().Msg("Waiting for cache sync")
		m.Informer.WaitForCacheSync(ctx.Done())
		return nil
	})

	g.Go(func(ctx context.Context) error {
		logger.Info().Msg("Starting informer")
		m.Informer.Start(ctx.Done())
		return nil
	})

	return g.Wait()
}
