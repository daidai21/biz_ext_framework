package ext_model

import "sync"

// ExtObj defines the value contract stored in ExtMap.
type ExtObj interface {
	Key() string
}

// ExtModel defines the behavior contract implemented by ExtMap.
type ExtModel[V ExtObj] interface {
	Get(key string) (V, bool)
	Set(value V)
	Del(key string) (V, bool)
	ForEach(fn func(value V))
}

// ExtMap is a generic, concurrency-safe map wrapper for business extensions.
// The zero value is ready to use.
type ExtMap[V ExtObj] struct {
	mu   sync.RWMutex
	data map[string]V
}

func (m *ExtMap[V]) Set(value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		m.data = make(map[string]V)
	}
	m.data[value.Key()] = value
}

func (m *ExtMap[V]) Get(key string) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		var zero V
		return zero, false
	}

	value, ok := m.data[key]
	return value, ok
}

func (m *ExtMap[V]) Del(key string) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		var zero V
		return zero, false
	}

	value, ok := m.data[key]
	if ok {
		delete(m.data, key)
	}

	return value, ok
}

func (m *ExtMap[V]) ForEach(fn func(value V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, value := range m.data {
		fn(value)
	}
}
