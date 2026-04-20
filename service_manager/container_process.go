package service_manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/biz_process"
	"github.com/daidai21/biz_ext_framework/ext_process"
)

var (
	ErrInvalidProcessName          = errors.New("invalid process name")
	ErrProcessNotFound             = errors.New("process not found")
	ErrInvalidExtProcessDefinition = errors.New("invalid ext process definition")
	ErrNilExtProcessTemplate       = errors.New("nil ext process template")
)

// ProcessContainer manages multiple named processes.
type ProcessContainer struct {
	mu        sync.RWMutex
	processes map[string]biz_process.Process
}

func NewProcessContainer() *ProcessContainer {
	return &ProcessContainer{}
}

func (c *ProcessContainer) Register(name string, process biz_process.Process) error {
	if name == "" {
		return ErrInvalidProcessName
	}
	if len(process.Layers) == 0 {
		return biz_process.ErrEmptyProcess
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.processes == nil {
		c.processes = make(map[string]biz_process.Process)
	}

	if process.Name == "" {
		process.Name = name
	}
	c.processes[name] = process
	return nil
}

func (c *ProcessContainer) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.processes, name)
}

func (c *ProcessContainer) Get(name string) (biz_process.Process, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	process, ok := c.processes[name]
	return process, ok
}

func (c *ProcessContainer) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.processes))
	for name := range c.processes {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (c *ProcessContainer) Run(ctx context.Context, name string) error {
	process, ok := c.Get(name)
	if !ok {
		return fmt.Errorf("%w: %q", ErrProcessNotFound, name)
	}
	return biz_process.RunProcess(ctx, process)
}

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
	return c.RegisterWithActionAndAppendType(definition, impl, ext_process.Append, ext_process.AppendAfter)
}

func (c *ExtProcessContainer[Impl, Input, Output]) RegisterWithAppendType(definition string, impl Impl, appendType ext_process.AppendType) error {
	return c.RegisterWithActionAndAppendType(definition, impl, ext_process.Append, appendType)
}

func (c *ExtProcessContainer[Impl, Input, Output]) RegisterWithAction(definition string, impl Impl, action ext_process.DefinitionAction) error {
	return c.RegisterWithActionAndAppendType(definition, impl, action, ext_process.AppendAfter)
}

func (c *ExtProcessContainer[Impl, Input, Output]) RegisterWithActionAndAppendType(definition string, impl Impl, action ext_process.DefinitionAction, appendType ext_process.AppendType) error {
	return c.ApplyWithAppendType(definition, []Impl{impl}, action, appendType)
}

func (c *ExtProcessContainer[Impl, Input, Output]) Apply(definition string, impls []Impl, action ext_process.DefinitionAction) error {
	return c.ApplyWithAppendType(definition, impls, action, ext_process.AppendAfter)
}

func (c *ExtProcessContainer[Impl, Input, Output]) ApplyWithAppendType(definition string, impls []Impl, action ext_process.DefinitionAction, appendType ext_process.AppendType) error {
	if definition == "" {
		return ErrInvalidExtProcessDefinition
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.impls == nil {
		c.impls = make(map[string][]Impl)
	}
	merged, err := ext_process.MergeImplementationsWithAppendType(c.impls[definition], impls, action, appendType)
	if err != nil {
		return err
	}
	c.impls[definition] = merged
	return nil
}

func (c *ExtProcessContainer[Impl, Input, Output]) Replace(definition string, impls []Impl) error {
	return c.Apply(definition, impls, ext_process.Overwrite)
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
