package testenv

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type DelegatingClient struct {
	client.Reader
	client.StatusClient
	client.Writer

	scheme *runtime.Scheme
	mapper meta.RESTMapper
}

// Scheme returns the scheme this client is using.
func (d *DelegatingClient) Scheme() *runtime.Scheme {
	return d.scheme
}

// RESTMapper returns the rest mapper this client is using.
func (d *DelegatingClient) RESTMapper() meta.RESTMapper {
	return d.mapper
}

type Manager struct {
	manager.Manager

	ctx    context.Context
	cancel context.CancelFunc
}

func (m *Manager) Initialize() (err error) {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	go func() {
		err = m.Manager.Start(m.ctx)
	}()

	m.Manager.GetCache().WaitForCacheSync(m.ctx)

	return
}

func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.ctx = nil
		m.cancel = nil
	}
}

func (m *Manager) GetClient() client.Client {
	c := m.Manager.GetClient()
	scheme := m.Manager.GetScheme()

	return &DelegatingClient{
		Reader: c,
		StatusClient: &recorderStatusClient{
			writer: c.Status(),
			scheme: scheme,
		},
		Writer: &recorderWriter{
			writer: c,
			scheme: scheme,
		},
		scheme: scheme,
		mapper: m.GetRESTMapper(),
	}
}

func (m *Manager) WaitForEvent(expected interface{}) bool {
	list := new(corev1.EventList)

	if err := m.GetClient().List(m.ctx, list); err != nil {
		panic(err)
	}

	for _, item := range list.Items {
		ok := matchEvent(EventData{
			Type:    item.Type,
			Reason:  item.Reason,
			Message: item.Message,
		}, expected)

		if ok {
			return true
		}
	}

	return false
}
