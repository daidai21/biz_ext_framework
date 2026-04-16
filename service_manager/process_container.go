package service_manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/biz_process"
)

var (
	ErrInvalidProcessName = errors.New("invalid process name")
	ErrProcessNotFound    = errors.New("process not found")
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
