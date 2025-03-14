package main

import "sync"

// Map type-safe sync map wrapper
type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Delete(key K) { m.m.Delete(key) }

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (m *Map[K, V]) Length() uint {
	count := uint(0)
	m.m.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func (m *Map[K, V]) Clear() {
	m.m.Clear()
}

func (m *Map[K, V]) Store(key K, value V) { m.m.Store(key, value) }
