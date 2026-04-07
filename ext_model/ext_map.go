package ext_model

import "sync"

// ExtObj defines the value contract stored in ExtMap.
type ExtObj interface {
	any
}

// ExtMap is a generic, concurrency-safe map wrapper for business extensions.
// The zero value is ready to use.
type ExtMap[V ExtObj] struct {
	mu   sync.RWMutex
	data map[string]V
}

func NewExtMap[V ExtObj](initial map[string]V) *ExtMap[V] {
	m := &ExtMap[V]{}
	if len(initial) > 0 {
		m.data = cloneMap(initial)
	}
	return m
}

func (m *ExtMap[V]) Set(key string, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureData()
	m.data[key] = value
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

func (m *ExtMap[V]) Has(key string) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *ExtMap[V]) Delete(key string) (V, bool) {
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

func (m *ExtMap[V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.data)
}

func (m *ExtMap[V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.data)
}

func (m *ExtMap[V]) LoadOrStore(key string, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureData()

	current, ok := m.data[key]
	if ok {
		return current, true
	}

	m.data[key] = value
	return value, false
}

func (m *ExtMap[V]) Range(fn func(key string, value V) bool) {
	snapshot := m.Clone()
	for key, value := range snapshot {
		if !fn(key, value) {
			return
		}
	}
}

func (m *ExtMap[V]) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

func (m *ExtMap[V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := make([]V, 0, len(m.data))
	for _, value := range m.data {
		values = append(values, value)
	}
	return values
}

func (m *ExtMap[V]) Merge(other map[string]V) {
	if len(other) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureData()
	for key, value := range other {
		m.data[key] = value
	}
}

func (m *ExtMap[V]) Clone() map[string]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return cloneMap(m.data)
}

func (m *ExtMap[V]) ensureData() {
	if m.data == nil {
		m.data = make(map[string]V)
	}
}

func cloneMap[V ExtObj](source map[string]V) map[string]V {
	if len(source) == 0 {
		return map[string]V{}
	}

	target := make(map[string]V, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
