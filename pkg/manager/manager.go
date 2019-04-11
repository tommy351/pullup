package manager

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned"
	"github.com/tommy351/pullup/pkg/client/informers/externalversions"
	"github.com/tommy351/pullup/pkg/group"
	"github.com/tommy351/pullup/pkg/k8s"
	"k8s.io/client-go/dynamic"
)

type Manager struct {
	Namespace string
	Client    versioned.Interface
	Informer  externalversions.SharedInformerFactory
	Dynamic   dynamic.Interface
}

func (m *Manager) Run(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	g := group.NewGroup(ctx)

	webhookHandler := &Handler{
		Kind:     k8s.Kind("Webhook"),
		MaxRetry: 5,
		Handler: &WebhookEventHandler{
			Client: m.Client,
		},
	}

	rsHandler := &Handler{
		Kind:     k8s.Kind("ResourceSet"),
		MaxRetry: 5,
		Handler: &ResourceSetEventHandler{
			Dynamic: m.Dynamic,
		},
	}

	m.Informer.Pullup().V1alpha1().Webhooks().Informer().AddEventHandler(webhookHandler)
	m.Informer.Pullup().V1alpha1().ResourceSets().Informer().AddEventHandler(rsHandler)

	g.Go(webhookHandler.Run)
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
