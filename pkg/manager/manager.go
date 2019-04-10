package manager

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/group"
	"github.com/tommy351/pullup/pkg/k8s"
)

type Manager struct {
	Client *k8s.Client
}

func (m *Manager) Run(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	informer := m.Client.NewInformer(ctx)
	g := group.NewGroup(ctx)

	rsHandler := &Handler{
		Kind:     v1alpha1.Kind("ResourceSet"),
		MaxRetry: 5,
		Handler: &ResourceSetEventHandler{
			Client: m.Client,
		},
	}

	informer.Pullup().V1alpha1().ResourceSets().Informer().AddEventHandler(rsHandler)

	g.Go(rsHandler.Run)

	g.Go(func(ctx context.Context) error {
		logger.Debug().Msg("Waiting for cache sync")
		informer.WaitForCacheSync(ctx.Done())
		return nil
	})

	g.Go(func(ctx context.Context) error {
		logger.Info().Msg("Starting informer")
		informer.Start(ctx.Done())
		return nil
	})

	return g.Wait()
}
