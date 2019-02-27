package cache

import "sync"

type Map struct {
	data  map[interface{}]interface{}
	mutex sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		data: map[interface{}]interface{}{},
	}
}

func (m *Map) Load(key interface{}) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	value, ok := m.data[key]
	return value, ok
}

func (m *Map) Store(key interface{}, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data[key] = value
}

func (m *Map) LoadOrStore(key interface{}, value func() interface{}) (interface{}, bool) {
	v, ok := m.Load(key)

	if ok {
		return v, ok
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	v = value()
	m.data[key] = v
	return v, false
}

func (m *Map) Delete(key interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)
}
