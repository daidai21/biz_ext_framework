package service_manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/ext_interceptor"
)

var (
	ErrInvalidInterceptorDefinition = errors.New("invalid interceptor definition")
	ErrNilInterceptorTemplate       = errors.New("nil interceptor template")
)

// InterceptorContainer manages ext_interceptor implementations grouped by definition key.
// One container binds one ext_interceptor template and executes implementations through it.
type InterceptorContainer[Impl any, Input any, Output any] struct {
	mu           sync.RWMutex
	template     ext_interceptor.Template[Impl, Input, Output]
	interceptors map[string][]Impl
}

func NewInterceptorContainer[Impl any, Input any, Output any](template ext_interceptor.Template[Impl, Input, Output]) (*InterceptorContainer[Impl, Input, Output], error) {
	if template == nil {
		return nil, ErrNilInterceptorTemplate
	}
	return &InterceptorContainer[Impl, Input, Output]{
		template: template,
	}, nil
}

func (c *InterceptorContainer[Impl, Input, Output]) Register(definition string, interceptor Impl) error {
	if definition == "" {
		return ErrInvalidInterceptorDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.interceptors == nil {
		c.interceptors = make(map[string][]Impl)
	}
	c.interceptors[definition] = append(c.interceptors[definition], interceptor)
	return nil
}

func (c *InterceptorContainer[Impl, Input, Output]) Replace(definition string, interceptors []Impl) error {
	if definition == "" {
		return ErrInvalidInterceptorDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.interceptors == nil {
		c.interceptors = make(map[string][]Impl)
	}
	c.interceptors[definition] = append([]Impl(nil), interceptors...)
	return nil
}

func (c *InterceptorContainer[Impl, Input, Output]) Remove(definition string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.interceptors, definition)
}

func (c *InterceptorContainer[Impl, Input, Output]) Interceptors(definition string) []Impl {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]Impl(nil), c.interceptors[definition]...)
}

func (c *InterceptorContainer[Impl, Input, Output]) Definitions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	definitions := make([]string, 0, len(c.interceptors))
	for definition := range c.interceptors {
		definitions = append(definitions, definition)
	}
	slices.Sort(definitions)
	return definitions
}

func (c *InterceptorContainer[Impl, Input, Output]) Execute(ctx context.Context, definition string, input Input, final ext_interceptor.Handler[Input, Output]) (Output, error) {
	if definition == "" {
		var zero Output
		return zero, ErrInvalidInterceptorDefinition
	}

	c.mu.RLock()
	template := c.template
	interceptors := append([]Impl(nil), c.interceptors[definition]...)
	c.mu.RUnlock()

	if template == nil {
		var zero Output
		return zero, ErrNilInterceptorTemplate
	}
	output, err := template(ctx, interceptors, input, final)
	if err != nil {
		var zero Output
		return zero, fmt.Errorf("interceptor definition %q execute failed: %w", definition, err)
	}
	return output, nil
}
