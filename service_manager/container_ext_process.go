package service_manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/ext_process"
)

var (
	ErrInvalidExtProcessDefinition = errors.New("invalid ext process definition")
	ErrNilExtProcessTemplate       = errors.New("nil ext process template")
)

// ExtProcessContainer manages ext_process implementations grouped by definition key.
// One container binds one ext_process template and executes implementations through it.
type ExtProcessContainer[Impl any, Input any, Output any] struct {
	mu       sync.RWMutex
	template ext_process.Template[Impl, Input, Output]
	impls    map[string][]Impl
}

func NewExtProcessContainer[Impl any, Input any, Output any](template ext_process.Template[Impl, Input, Output]) (*ExtProcessContainer[Impl, Input, Output], error) {
	if template == nil {
		return nil, ErrNilExtProcessTemplate
	}
	return &ExtProcessContainer[Impl, Input, Output]{
		template: template,
	}, nil
}

func (c *ExtProcessContainer[Impl, Input, Output]) Register(definition string, impl Impl) error {
	if definition == "" {
		return ErrInvalidExtProcessDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.impls == nil {
		c.impls = make(map[string][]Impl)
	}
	c.impls[definition] = append(c.impls[definition], impl)
	return nil
}

func (c *ExtProcessContainer[Impl, Input, Output]) Replace(definition string, impls []Impl) error {
	if definition == "" {
		return ErrInvalidExtProcessDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.impls == nil {
		c.impls = make(map[string][]Impl)
	}
	c.impls[definition] = append([]Impl(nil), impls...)
	return nil
}

func (c *ExtProcessContainer[Impl, Input, Output]) Remove(definition string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.impls, definition)
}

func (c *ExtProcessContainer[Impl, Input, Output]) Implementations(definition string) []Impl {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]Impl(nil), c.impls[definition]...)
}

func (c *ExtProcessContainer[Impl, Input, Output]) Definitions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	definitions := make([]string, 0, len(c.impls))
	for definition := range c.impls {
		definitions = append(definitions, definition)
	}
	slices.Sort(definitions)
	return definitions
}

func (c *ExtProcessContainer[Impl, Input, Output]) Execute(ctx context.Context, definition string, input Input, mode ext_process.Mode) ([]Output, error) {
	if definition == "" {
		return nil, ErrInvalidExtProcessDefinition
	}

	c.mu.RLock()
	template := c.template
	impls := append([]Impl(nil), c.impls[definition]...)
	c.mu.RUnlock()

	if template == nil {
		return nil, ErrNilExtProcessTemplate
	}
	results, err := template(ctx, impls, input, mode)
	if err != nil {
		return nil, fmt.Errorf("ext process definition %q execute failed: %w", definition, err)
	}
	return results, nil
}
