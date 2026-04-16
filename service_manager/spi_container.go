package service_manager

import (
	"errors"
	"slices"
	"sync"
)

var (
	ErrInvalidSPIDefinition = errors.New("invalid spi definition")
)

// SPIContainer manages implementations grouped by SPI definition key.
type SPIContainer[Impl any] struct {
	mu    sync.RWMutex
	impls map[string][]Impl
}

func NewSPIContainer[Impl any]() *SPIContainer[Impl] {
	return &SPIContainer[Impl]{}
}

func (c *SPIContainer[Impl]) Register(definition string, impl Impl) error {
	if definition == "" {
		return ErrInvalidSPIDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.impls == nil {
		c.impls = make(map[string][]Impl)
	}
	c.impls[definition] = append(c.impls[definition], impl)
	return nil
}

func (c *SPIContainer[Impl]) Replace(definition string, impls []Impl) error {
	if definition == "" {
		return ErrInvalidSPIDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.impls == nil {
		c.impls = make(map[string][]Impl)
	}
	c.impls[definition] = append([]Impl(nil), impls...)
	return nil
}

func (c *SPIContainer[Impl]) Remove(definition string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.impls, definition)
}

func (c *SPIContainer[Impl]) Implementations(definition string) []Impl {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]Impl(nil), c.impls[definition]...)
}

func (c *SPIContainer[Impl]) Definitions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	definitions := make([]string, 0, len(c.impls))
	for definition := range c.impls {
		definitions = append(definitions, definition)
	}
	slices.Sort(definitions)
	return definitions
}
