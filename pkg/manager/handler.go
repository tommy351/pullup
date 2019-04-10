package manager

import (
	"context"

	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
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
	Kind     schema.GroupVersionKind
	MaxRetry int
	Handler  EventHandler

	updateQueue, deleteQueue workqueue.RateLimitingInterface
}

func (h *Handler) Run(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	h.updateQueue = h.newQueue("update")
	h.deleteQueue = h.newQueue("delete")

	go h.runQueue(ctx, h.updateQueue, h.Handler.OnUpdate)
	go h.runQueue(ctx, h.deleteQueue, h.Handler.OnDelete)

	<-ctx.Done()

	logger.Debug().Str("kind", h.Kind.String()).Msg("Shutting down handler")
	h.updateQueue.ShutDown()
	h.deleteQueue.ShutDown()

	return nil
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

func (h *Handler) newQueue(event string) workqueue.RateLimitingInterface {
	return workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), h.Kind.String()+", Event="+event)
}

func (h *Handler) runQueue(ctx context.Context, queue workqueue.RateLimitingInterface, handler func(ctx context.Context, obj interface{}) error) {
	process := func() bool {
		item, shutdown := queue.Get()

		if shutdown {
			return false
		}

		defer queue.Done(item)

		if obj, ok := item.(schema.ObjectKind); ok && obj.GroupVersionKind().Empty() {
			obj.SetGroupVersionKind(h.Kind)
		}

		key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(item)
		retry := queue.NumRequeues(item)
		logger := zerolog.Ctx(ctx).With().
			Dict("task", zerolog.Dict().
				Str("key", key).
				Int("retry", retry).
				Int("maxRetry", h.MaxRetry)).
			Logger()

		ctx := logger.WithContext(ctx)
		err := handler(ctx, item)

		if err == nil {
			logger.Debug().Msg("Task is done")
			queue.Forget(item)
			return true
		}

		if retry < h.MaxRetry {
			logger.Debug().Msg("Task is requeued")
			queue.AddRateLimited(item)
			return true
		}

		logger.Error().Stack().Err(err).Msg("Task is dropped")
		queue.Forget(item)
		return true
	}

	for process() {
	}
}
