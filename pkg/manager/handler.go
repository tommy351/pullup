package manager

import (
	"context"

	"github.com/rs/zerolog"
	"k8s.io/client-go/util/workqueue"
)

type resourceVersionGetter interface {
	GetResourceVersion() string
}

type EventHandler interface {
	OnUpdate(ctx context.Context, obj interface{}) error
	OnDelete(ctx context.Context, obj interface{}) error
}

type Handler struct {
	name                     string
	updateQueue, deleteQueue workqueue.Interface
}

func NewHandler(ctx context.Context, name string, handler EventHandler) *Handler {
	h := new(Handler)
	h.name = name
	h.updateQueue = h.newQueue(name + "/update")
	h.deleteQueue = h.newQueue(name + "/delete")

	go h.waitForShutdown(ctx)
	go h.runQueue(ctx, h.updateQueue, handler.OnUpdate)
	go h.runQueue(ctx, h.deleteQueue, handler.OnDelete)

	return h
}

func (h *Handler) OnAdd(obj interface{}) {
	h.updateQueue.Add(obj)
}

func (h *Handler) OnDelete(obj interface{}) {
	h.deleteQueue.Add(obj)
}

func (h *Handler) OnUpdate(oldObj, newObj interface{}) {
	oldMeta := oldObj.(resourceVersionGetter)
	newMeta := newObj.(resourceVersionGetter)

	if oldMeta.GetResourceVersion() == newMeta.GetResourceVersion() {
		return
	}

	h.updateQueue.Add(newObj)
}

func (h *Handler) newQueue(name string) workqueue.Interface {
	return workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name)
}

func (h *Handler) runQueue(ctx context.Context, queue workqueue.Interface, handler func(ctx context.Context, obj interface{}) error) {
	logger := zerolog.Ctx(ctx).With().Str("name", h.name).Logger()

	for {
		item, shutdown := queue.Get()

		if shutdown {
			return
		}

		if err := handler(ctx, item); err != nil {
			logger.Error().Stack().Err(err).Msg("Event handler error")
		} else {
			queue.Done(item)
		}
	}
}

func (h *Handler) waitForShutdown(ctx context.Context) {
	logger := zerolog.Ctx(ctx)

	<-ctx.Done()

	logger.Debug().Str("name", h.name).Msg("Shutting down handler")
	h.updateQueue.ShutDown()
	h.deleteQueue.ShutDown()
}
