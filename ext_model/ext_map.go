package ext_model

import "sync"

// ExtMap is a generic, concurrency-safe map wrapper for business extensions.
// The zero value is ready to use.
type ExtMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewExtMap[K comparable, V any](initial map[K]V) *ExtMap[K, V] {
	m := &ExtMap[K, V]{}
	if len(initial) > 0 {
		m.data = cloneMap(initial)
	}
	return m
}

func (m *ExtMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureData()
	m.data[key] = value
}

func (m *ExtMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		var zero V
		return zero, false
	}

	value, ok := m.data[key]
	return value, ok
}

func (m *ExtMap[K, V]) Has(key K) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *ExtMap[K, V]) Delete(key K) (V, bool) {
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

func (m *ExtMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.data)
}

func (m *ExtMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.data)
}

func (m *ExtMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
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

func (m *ExtMap[K, V]) Range(fn func(key K, value V) bool) {
	snapshot := m.Clone()
	for key, value := range snapshot {
		if !fn(key, value) {
			return
		}
	}
}

func (m *ExtMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]K, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

func (m *ExtMap[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := make([]V, 0, len(m.data))
	for _, value := range m.data {
		values = append(values, value)
	}
	return values
}

func (m *ExtMap[K, V]) Merge(other map[K]V) {
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

func (m *ExtMap[K, V]) Clone() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return cloneMap(m.data)
}

func (m *ExtMap[K, V]) ensureData() {
	if m.data == nil {
		m.data = make(map[K]V)
	}
}

func cloneMap[K comparable, V any](source map[K]V) map[K]V {
	if len(source) == 0 {
		return map[K]V{}
	}

	target := make(map[K]V, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
