package service_manager

import (
	"context"

	"github.com/daidai21/biz_ext_framework/biz_component"
)

// ComponentContainer manages IOC-style business components for global/session singletons.
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

func (c *ComponentContainer) RegisterAnyIn(name string, scope biz_component.Scope, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error {
	return c.container.RegisterAnyIn(name, scope, namespace, provider)
}

func (c *ComponentContainer) RegisterGlobalIn(name string, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error {
	return c.container.RegisterAnyIn(name, biz_component.GlobalScope, namespace, provider)
}

func (c *ComponentContainer) RegisterSessionIn(name string, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error {
	return c.container.RegisterAnyIn(name, biz_component.SessionScope, namespace, provider)
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

func (c *ComponentContainer) GlobalObject(name string) (any, bool) {
	return c.container.GlobalObjectAny(name)
}

func (c *ComponentContainer) SessionObject(sessionID, name string) (any, bool) {
	return c.container.SessionObjectAny(sessionID, name)
}

func (c *ComponentContainer) GlobalObjects() map[string]any {
	return c.container.GlobalObjects()
}

func (c *ComponentContainer) SessionObjects(sessionID string) map[string]any {
	return c.container.SessionObjects(sessionID)
}

func (c *ComponentContainer) GlobalNames() []string {
	return c.container.GlobalNames()
}

func (c *ComponentContainer) SessionNames(sessionID string) []string {
	return c.container.SessionNames(sessionID)
}

func (c *ComponentContainer) DeleteGlobal(name string) {
	c.container.DeleteGlobal(name)
}

func (c *ComponentContainer) DeleteSessionObject(sessionID, name string) {
	c.container.DeleteSessionObject(sessionID, name)
}

func (c *ComponentContainer) ClearSession(sessionID string) {
	c.container.ClearSession(sessionID)
}
