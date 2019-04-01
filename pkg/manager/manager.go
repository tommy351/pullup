package manager

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/k8s"
)

type Manager struct {
	Client k8s.Client
}

func (m *Manager) Run(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	informer := m.Client.NewInformer(ctx)

	informer.Pullup().V1alpha1().ResourceSets().Informer().AddEventHandler(NewHandler(ctx, "ResourceSet", &ResourceSetEventHandler{
		Client: m.Client,
	}))

	logger.Info().Msg("Waiting for cache sync")
	informer.WaitForCacheSync(ctx.Done())

	logger.Info().Msg("Starting informer")
	informer.Start(ctx.Done())
	return nil
}
