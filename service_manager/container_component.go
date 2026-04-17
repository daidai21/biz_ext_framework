package service_manager

import (
	"context"

	"github.com/daidai21/biz_ext_framework/biz_component"
)

// ComponentContainer manages IOC-style business components for both
// service scope and session scope.
type ComponentContainer struct {
	container    *biz_component.Container
	ctxContainer *CtxContainer
}

func NewComponentContainer(ctxContainer *CtxContainer) *ComponentContainer {
	return &ComponentContainer{
		container:    biz_component.NewContainer(),
		ctxContainer: ctxContainer,
	}
}

func (c *ComponentContainer) Container() *biz_component.Container {
	return c.container
}

func (c *ComponentContainer) RegisterAny(name string, scope biz_component.Scope, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error {
	return c.container.RegisterAny(name, scope, provider)
}

func (c *ComponentContainer) ResolveAny(ctx context.Context, name string) (any, error) {
	return c.container.ResolveAny(ctx, name)
}

func (c *ComponentContainer) ResolveAnyInSession(ctx context.Context, sessionID, name string) (any, error) {
	if c.ctxContainer != nil {
		sessionCtx, err := c.ctxContainer.WithSession(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return c.container.ResolveAny(sessionCtx, name)
	}
	return c.container.ResolveAny(ctx, name)
}

func (c *ComponentContainer) ServiceObject(name string) (any, bool) {
	return c.container.ServiceObjectAny(name)
}

func (c *ComponentContainer) SessionObject(sessionID, name string) (any, bool) {
	return c.container.SessionObjectAny(sessionID, name)
}

func (c *ComponentContainer) ServiceObjects() map[string]any {
	return c.container.ServiceObjects()
}

func (c *ComponentContainer) SessionObjects(sessionID string) map[string]any {
	return c.container.SessionObjects(sessionID)
}

func (c *ComponentContainer) ServiceNames() []string {
	return c.container.ServiceNames()
}

func (c *ComponentContainer) SessionNames(sessionID string) []string {
	return c.container.SessionNames(sessionID)
}

func (c *ComponentContainer) DeleteService(name string) {
	c.container.DeleteService(name)
}

func (c *ComponentContainer) DeleteSessionObject(sessionID, name string) {
	c.container.DeleteSessionObject(sessionID, name)
}

func (c *ComponentContainer) ClearSession(sessionID string) {
	c.container.ClearSession(sessionID)
}
