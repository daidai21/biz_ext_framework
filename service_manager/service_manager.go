package service_manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const (
	IdentityContainerName = "identity_container"
	ProcessContainerName  = "process_container"
	ModelContainerName    = "model_container"
)

var (
	ErrInvalidServiceManagerName = errors.New("invalid service manager name")
	ErrInvalidContainerName      = errors.New("invalid container name")
	ErrNilContainer              = errors.New("nil container")
	ErrNilStartupCheck           = errors.New("nil startup check")
	ErrInvalidLifecycleName      = errors.New("invalid lifecycle name")
	ErrNilLifecycle              = errors.New("nil lifecycle")
	ErrContainerNotFound         = errors.New("container not found")
	ErrServiceManagerStarted     = errors.New("service manager already started")
	ErrServiceManagerNotStarted  = errors.New("service manager not started")
	ErrServiceManagerStopped     = errors.New("service manager already stopped")
	ErrServiceManagerBusy        = errors.New("service manager is transitioning")
)

type ServiceManagerState string

const (
	ServiceManagerStateReady   ServiceManagerState = "READY"
	ServiceManagerStateStarted ServiceManagerState = "STARTED"
	ServiceManagerStateStopped ServiceManagerState = "STOPPED"
)

// StartupCheck validates manager readiness before startup.
type StartupCheck func(ctx context.Context, manager *ServiceManager) error

// Lifecycle represents one managed service resource.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type namedLifecycle struct {
	name      string
	lifecycle Lifecycle
}

// ServiceManager manages initialized containers and service instance lifecycle.
type ServiceManager struct {
	mu sync.Mutex

	name  string
	state ServiceManagerState
	busy  bool

	identityContainer *IdentityContainer
	processContainer  *ProcessContainer
	modelContainer    *ModelContainer
	containers        map[string]any

	startupChecks []StartupCheck
	lifecycles    []namedLifecycle
}

func (m *ServiceManager) Name() string {
	return m.name
}

func (m *ServiceManager) State() ServiceManagerState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

func (m *ServiceManager) IdentityContainer() *IdentityContainer {
	return m.identityContainer
}

func (m *ServiceManager) ProcessContainer() *ProcessContainer {
	return m.processContainer
}

func (m *ServiceManager) ModelContainer() *ModelContainer {
	return m.modelContainer
}

func (m *ServiceManager) Container(name string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	container, ok := m.containers[name]
	return container, ok
}

func (m *ServiceManager) MustContainer(name string) (any, error) {
	container, ok := m.Container(name)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrContainerNotFound, name)
	}
	return container, nil
}

func (m *ServiceManager) Check(ctx context.Context) error {
	m.mu.Lock()
	checks := append([]StartupCheck(nil), m.startupChecks...)
	m.mu.Unlock()

	for i, check := range checks {
		if err := check(ctx, m); err != nil {
			return fmt.Errorf("startup check[%d] failed: %w", i, err)
		}
	}
	return nil
}

func (m *ServiceManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return ErrServiceManagerBusy
	}
	switch m.state {
	case ServiceManagerStateStarted:
		m.mu.Unlock()
		return ErrServiceManagerStarted
	case ServiceManagerStateStopped:
		m.mu.Unlock()
		return ErrServiceManagerStopped
	}
	m.busy = true
	lifecycles := append([]namedLifecycle(nil), m.lifecycles...)
	m.mu.Unlock()

	if err := m.Check(ctx); err != nil {
		m.mu.Lock()
		m.busy = false
		m.mu.Unlock()
		return err
	}

	started := make([]namedLifecycle, 0, len(lifecycles))
	for i, item := range lifecycles {
		if err := item.lifecycle.Start(ctx); err != nil {
			m.mu.Lock()
			m.busy = false
			m.mu.Unlock()
			rollbackErr := stopLifecycles(ctx, started)
			startErr := fmt.Errorf("lifecycle[%d] %q start failed: %w", i, item.name, err)
			if rollbackErr != nil {
				return errors.Join(startErr, fmt.Errorf("startup rollback failed: %w", rollbackErr))
			}
			return startErr
		}
		started = append(started, item)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	defer func() {
		m.busy = false
	}()
	if m.state == ServiceManagerStateStopped {
		return ErrServiceManagerStopped
	}
	m.state = ServiceManagerStateStarted
	return nil
}

func (m *ServiceManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return ErrServiceManagerBusy
	}
	switch m.state {
	case ServiceManagerStateReady:
		m.mu.Unlock()
		return ErrServiceManagerNotStarted
	case ServiceManagerStateStopped:
		m.mu.Unlock()
		return ErrServiceManagerStopped
	}
	m.busy = true
	lifecycles := append([]namedLifecycle(nil), m.lifecycles...)
	m.mu.Unlock()

	err := stopLifecycles(ctx, lifecycles)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.busy = false
	m.state = ServiceManagerStateStopped
	return err
}

func stopLifecycles(ctx context.Context, lifecycles []namedLifecycle) error {
	var errs []error
	for i := len(lifecycles) - 1; i >= 0; i-- {
		item := lifecycles[i]
		if err := item.lifecycle.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("lifecycle %q stop failed: %w", item.name, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
