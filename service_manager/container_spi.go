package service_manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/ext_spi"
)

var (
	ErrInvalidSPIDefinition = errors.New("invalid spi definition")
	ErrNilSPITemplate       = errors.New("nil spi template")
)

// SPIContainer manages ext_spi implementations grouped by SPI definition key.
// One container binds one ext_spi template and executes implementations through it.
type SPIContainer[Impl any, Input any, Output any] struct {
	mu       sync.RWMutex
	template ext_spi.Template[Impl, Input, Output]
	impls    map[string][]Impl
}

func NewSPIContainer[Impl any, Input any, Output any](template ext_spi.Template[Impl, Input, Output]) (*SPIContainer[Impl, Input, Output], error) {
	if template == nil {
		return nil, ErrNilSPITemplate
	}
	return &SPIContainer[Impl, Input, Output]{
		template: template,
	}, nil
}

func (c *SPIContainer[Impl, Input, Output]) Register(definition string, impl Impl) error {
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

func (c *SPIContainer[Impl, Input, Output]) Replace(definition string, impls []Impl) error {
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

func (c *SPIContainer[Impl, Input, Output]) Remove(definition string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.impls, definition)
}

func (c *SPIContainer[Impl, Input, Output]) Implementations(definition string) []Impl {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]Impl(nil), c.impls[definition]...)
}

func (c *SPIContainer[Impl, Input, Output]) Definitions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	definitions := make([]string, 0, len(c.impls))
	for definition := range c.impls {
		definitions = append(definitions, definition)
	}
	slices.Sort(definitions)
	return definitions
}

func (c *SPIContainer[Impl, Input, Output]) Execute(ctx context.Context, definition string, input Input, mode ext_spi.Mode) ([]Output, error) {
	if definition == "" {
		return nil, ErrInvalidSPIDefinition
	}

	c.mu.RLock()
	template := c.template
	impls := append([]Impl(nil), c.impls[definition]...)
	c.mu.RUnlock()

	if template == nil {
		return nil, ErrNilSPITemplate
	}
	results, err := template(ctx, impls, input, mode)
	if err != nil {
		return nil, fmt.Errorf("spi definition %q execute failed: %w", definition, err)
	}
	return results, nil
}
