package testenv

import (
	"reflect"
	"sync"
	"time"

	"github.com/eapache/channels"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
)

type EventData struct {
	Type    string
	Reason  string
	Message string
}

type EventWatcher struct {
	ch     *channels.InfiniteChannel
	mutex  sync.RWMutex
	closed bool
}

func NewEventWatcher() *EventWatcher {
	return &EventWatcher{
		ch: channels.NewInfiniteChannel(),
	}
}

func (e *EventWatcher) Shutdown() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.ch.Close()
	e.closed = true
}

func (e *EventWatcher) sendEvent(event interface{}) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if !e.closed {
		e.ch.In() <- event
	}
}

func (e *EventWatcher) WatchEvent(event *corev1.Event) {
	e.sendEvent(EventData{
		Type:    event.Type,
		Reason:  event.Reason,
		Message: event.Message,
	})
}

func (e *EventWatcher) WaitForEventTimeout(expected interface{}, timeout time.Duration) bool {
	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			return false

		case input, open := <-e.ch.Out():
			if !open {
				return false
			}

			e.sendEvent(input)

			switch e := expected.(type) {
			case types.GomegaMatcher:
				if ok, _ := e.Match(input); ok {
					return true
				}

			default:
				if reflect.DeepEqual(input, expected) {
					return true
				}
			}
		}
	}
}
