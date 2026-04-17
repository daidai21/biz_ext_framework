package service_manager

import (
	"fmt"

	"github.com/daidai21/biz_ext_framework/biz_process"
)

type processBuildConfig struct {
	name    string
	process biz_process.Process
}

type whitelistBuildConfig struct {
	rpcMethod   string
	allowedKeys []string
}

type containerBuildConfig struct {
	name      string
	container any
}

// ServiceManagerBuilder initializes standard containers, registers resources,
// and builds one service manager instance.
type ServiceManagerBuilder struct {
	name string

	identityContainer  *IdentityContainer
	processContainer   *ProcessContainer
	modelContainer     *ModelContainer
	ctxContainer       *CtxContainer
	componentContainer *ComponentContainer
	observation        *ObservationContainer

	identityScopes []string
	processes      []processBuildConfig
	whitelists     []whitelistBuildConfig
	containers     []containerBuildConfig
	startupChecks  []StartupCheck
	lifecycles     []namedLifecycle

	err error
}

func NewServiceManagerBuilder(name string) *ServiceManagerBuilder {
	return &ServiceManagerBuilder{name: name}
}

func (b *ServiceManagerBuilder) WithIdentityContainer(container *IdentityContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.identityContainer = container
	return b
}

func (b *ServiceManagerBuilder) WithProcessContainer(container *ProcessContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.processContainer = container
	return b
}

func (b *ServiceManagerBuilder) WithModelContainer(container *ModelContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.modelContainer = container
	return b
}

func (b *ServiceManagerBuilder) WithCtxContainer(container *CtxContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.ctxContainer = container
	return b
}

func (b *ServiceManagerBuilder) WithComponentContainer(container *ComponentContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.componentContainer = container
	return b
}

func (b *ServiceManagerBuilder) WithObservationContainer(container *ObservationContainer) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.observation = container
	return b
}

func (b *ServiceManagerBuilder) WithIdentityScopes(scopes ...string) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	b.identityScopes = append(b.identityScopes, scopes...)
	return b
}

func (b *ServiceManagerBuilder) WithProcess(name string, process biz_process.Process) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	b.processes = append(b.processes, processBuildConfig{name: name, process: process})
	return b
}

func (b *ServiceManagerBuilder) WithModelWhitelist(rpcMethod string, allowedKeys ...string) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	b.whitelists = append(b.whitelists, whitelistBuildConfig{
		rpcMethod:   rpcMethod,
		allowedKeys: append([]string(nil), allowedKeys...),
	})
	return b
}

func (b *ServiceManagerBuilder) WithContainer(name string, container any) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if name == "" {
		b.err = ErrInvalidContainerName
		return b
	}
	if container == nil {
		b.err = ErrNilContainer
		return b
	}
	b.containers = append(b.containers, containerBuildConfig{name: name, container: container})
	return b
}

func (b *ServiceManagerBuilder) WithStartupCheck(check StartupCheck) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if check == nil {
		b.err = ErrNilStartupCheck
		return b
	}
	b.startupChecks = append(b.startupChecks, check)
	return b
}

func (b *ServiceManagerBuilder) WithLifecycle(name string, lifecycle Lifecycle) *ServiceManagerBuilder {
	if b.err != nil {
		return b
	}
	if name == "" {
		b.err = ErrInvalidLifecycleName
		return b
	}
	if lifecycle == nil {
		b.err = ErrNilLifecycle
		return b
	}
	b.lifecycles = append(b.lifecycles, namedLifecycle{name: name, lifecycle: lifecycle})
	return b
}

func (b *ServiceManagerBuilder) Build() (*ServiceManager, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.name == "" {
		return nil, ErrInvalidServiceManagerName
	}

	identityContainer := b.identityContainer
	if identityContainer == nil {
		identityContainer = &IdentityContainer{}
	}
	processContainer := b.processContainer
	if processContainer == nil {
		processContainer = NewProcessContainer()
	}
	modelContainer := b.modelContainer
	if modelContainer == nil {
		modelContainer = NewModelContainer()
	}
	ctxContainer := b.ctxContainer
	if ctxContainer == nil {
		ctxContainer = NewCtxContainer()
	}
	componentContainer := b.componentContainer
	if componentContainer == nil {
		componentContainer = NewComponentContainer(ctxContainer)
	}
	observationContainer := b.observation
	if observationContainer == nil {
		observationContainer = NewObservationContainer()
	}

	for _, scope := range b.identityScopes {
		if err := identityContainer.AllowScope(scope); err != nil {
			return nil, err
		}
	}
	for _, item := range b.processes {
		if err := processContainer.Register(item.name, item.process); err != nil {
			return nil, fmt.Errorf("register process %q failed: %w", item.name, err)
		}
	}
	for _, item := range b.whitelists {
		if err := modelContainer.SetWhitelist(item.rpcMethod, item.allowedKeys); err != nil {
			return nil, fmt.Errorf("set model whitelist for %q failed: %w", item.rpcMethod, err)
		}
	}

	containers := map[string]any{
		IdentityContainerName:    identityContainer,
		ProcessContainerName:     processContainer,
		ModelContainerName:       modelContainer,
		CtxContainerName:         ctxContainer,
		ComponentContainerName:   componentContainer,
		ObservationContainerName: observationContainer,
	}
	for _, item := range b.containers {
		if _, exists := containers[item.name]; exists {
			return nil, fmt.Errorf("container %q already exists", item.name)
		}
		containers[item.name] = item.container
	}

	return &ServiceManager{
		name:               b.name,
		state:              ServiceManagerStateReady,
		identityContainer:  identityContainer,
		processContainer:   processContainer,
		modelContainer:     modelContainer,
		ctxContainer:       ctxContainer,
		componentContainer: componentContainer,
		observation:        observationContainer,
		containers:         containers,
		startupChecks:      append([]StartupCheck(nil), b.startupChecks...),
		lifecycles:         append([]namedLifecycle(nil), b.lifecycles...),
	}, nil
}
