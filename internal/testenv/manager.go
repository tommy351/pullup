package testenv

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager

	stopCh chan struct{}
}

func (m *Manager) Initialize() (err error) {
	m.stopCh = make(chan struct{})

	go func() {
		err = m.Manager.Start(m.stopCh)
	}()

	return
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) WaitForSync() {
	m.Manager.GetCache().WaitForCacheSync(m.stopCh)
}

func (m *Manager) GetClient() client.Client {
	c := m.Manager.GetClient()

	return &client.DelegatingClient{
		Reader:       c,
		StatusClient: c,
		Writer: &recorderWriter{
			writer: c,
			scheme: m.Manager.GetScheme(),
		},
	}
}
