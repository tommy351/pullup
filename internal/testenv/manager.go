package testenv

import (
	"time"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager

	EventBroadcaster record.EventBroadcaster

	stopCh       chan struct{}
	eventWatcher *EventWatcher
}

func (m *Manager) Initialize() (err error) {
	m.stopCh = make(chan struct{})
	m.eventWatcher = NewEventWatcher()

	go func() {
		err = m.Manager.Start(m.stopCh)
	}()

	m.WaitForSync()
	m.EventBroadcaster.StartEventWatcher(m.eventWatcher.WatchEvent)

	return
}

func (m *Manager) Stop() {
	close(m.stopCh)
	m.EventBroadcaster.Shutdown()
	m.eventWatcher.Shutdown()
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

func (m *Manager) WaitForEvent(expected interface{}) bool {
	return m.eventWatcher.WaitForEventTimeout(expected, time.Second)
}
