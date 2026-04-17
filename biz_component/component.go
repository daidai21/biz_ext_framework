package biz_component

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

var (
	ErrInvalidComponentName = errors.New("invalid component name")
	ErrNilProvider          = errors.New("nil component provider")
	ErrComponentNotFound    = errors.New("component not found")
	ErrSessionRequired      = errors.New("session is required")
	ErrCircularDependency   = errors.New("circular component dependency")
)

type Scope string

const (
	ServiceScope Scope = "SERVICE"
	SessionScope Scope = "SESSION"
)

type Provider func(ctx context.Context, resolver Resolver) (any, error)

type Resolver interface {
	Resolve(ctx context.Context, name string) (any, error)
}

type definition struct {
	scope    Scope
	provider Provider
}

type inflightBuild struct {
	wait  chan struct{}
	value any
	err   error
}

type resolverPathContextKey struct{}

// Container is a concurrency-safe IOC container supporting both service-level
// and session-level object management.
type Container struct {
	mu sync.Mutex

	definitions map[string]definition

	serviceObjects map[string]any
	sessionObjects map[string]map[string]any
	inflight       map[string]*inflightBuild
}

var _ Resolver = (*Container)(nil)

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Register(name string, scope Scope, provider Provider) error {
	if name == "" {
		return ErrInvalidComponentName
	}
	if provider == nil {
		return ErrNilProvider
	}
	if scope != ServiceScope && scope != SessionScope {
		return fmt.Errorf("unsupported component scope: %q", scope)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.definitions == nil {
		c.definitions = make(map[string]definition)
	}
	c.definitions[name] = definition{
		scope:    scope,
		provider: provider,
	}
	return nil
}

func (c *Container) RegisterService(name string, provider Provider) error {
	return c.Register(name, ServiceScope, provider)
}

func (c *Container) RegisterSession(name string, provider Provider) error {
	return c.Register(name, SessionScope, provider)
}

func (c *Container) Resolve(ctx context.Context, name string) (any, error) {
	if name == "" {
		return nil, ErrInvalidComponentName
	}

	c.mu.Lock()
	definition, ok := c.definitions[name]
	if !ok {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: %q", ErrComponentNotFound, name)
	}

	key, sessionID, err := c.resolveKey(ctx, name, definition.scope)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	if value, ok := c.cachedObjectLocked(definition.scope, sessionID, name); ok {
		c.mu.Unlock()
		return value, nil
	}

	if hasCycle(ctx, name) {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: %s", ErrCircularDependency, strings.Join(append(resolverPathFromContext(ctx), name), " -> "))
	}

	if inflight := c.getInflightLocked(key); inflight != nil {
		c.mu.Unlock()
		<-inflight.wait
		return inflight.value, inflight.err
	}

	inflight := &inflightBuild{wait: make(chan struct{})}
	if c.inflight == nil {
		c.inflight = make(map[string]*inflightBuild)
	}
	c.inflight[key] = inflight
	c.mu.Unlock()

	resolveCtx := withResolverPath(ctx, name)
	value, buildErr := definition.provider(resolveCtx, c)

	c.mu.Lock()
	if buildErr == nil {
		c.storeObjectLocked(definition.scope, sessionID, name, value)
	}
	delete(c.inflight, key)
	inflight.value = value
	inflight.err = buildErr
	close(inflight.wait)
	c.mu.Unlock()

	return value, buildErr
}

func (c *Container) ServiceObject(name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cachedObjectLocked(ServiceScope, "", name)
}

func (c *Container) SessionObject(sessionID, name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cachedObjectLocked(SessionScope, sessionID, name)
}

func (c *Container) ServiceObjects() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	copied := make(map[string]any, len(c.serviceObjects))
	for name, value := range c.serviceObjects {
		copied[name] = value
	}
	return copied
}

func (c *Container) SessionObjects(sessionID string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	source := c.sessionObjects[sessionID]
	copied := make(map[string]any, len(source))
	for name, value := range source {
		copied[name] = value
	}
	return copied
}

func (c *Container) ServiceNames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	names := make([]string, 0, len(c.serviceObjects))
	for name := range c.serviceObjects {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (c *Container) SessionNames(sessionID string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	names := make([]string, 0, len(c.sessionObjects[sessionID]))
	for name := range c.sessionObjects[sessionID] {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (c *Container) DeleteService(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.serviceObjects, name)
}

func (c *Container) DeleteSessionObject(sessionID, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionObjects == nil {
		return
	}
	objects := c.sessionObjects[sessionID]
	delete(objects, name)
	if len(objects) == 0 {
		delete(c.sessionObjects, sessionID)
	}
}

func (c *Container) ClearSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.sessionObjects, sessionID)
}

func (c *Container) resolveKey(ctx context.Context, name string, scope Scope) (key string, sessionID string, err error) {
	if scope == ServiceScope {
		return "service:" + name, "", nil
	}

	session, ok := biz_ctx.BizSessionFromContext(ctx)
	if !ok {
		return "", "", ErrSessionRequired
	}
	sessionID = session.BizSessionId()
	if sessionID == "" {
		return "", "", ErrSessionRequired
	}
	return "session:" + sessionID + ":" + name, sessionID, nil
}

func (c *Container) cachedObjectLocked(scope Scope, sessionID, name string) (any, bool) {
	switch scope {
	case ServiceScope:
		if c.serviceObjects == nil {
			return nil, false
		}
		value, ok := c.serviceObjects[name]
		return value, ok
	case SessionScope:
		if c.sessionObjects == nil {
			return nil, false
		}
		objects := c.sessionObjects[sessionID]
		value, ok := objects[name]
		return value, ok
	default:
		return nil, false
	}
}

func (c *Container) storeObjectLocked(scope Scope, sessionID, name string, value any) {
	switch scope {
	case ServiceScope:
		if c.serviceObjects == nil {
			c.serviceObjects = make(map[string]any)
		}
		c.serviceObjects[name] = value
	case SessionScope:
		if c.sessionObjects == nil {
			c.sessionObjects = make(map[string]map[string]any)
		}
		if c.sessionObjects[sessionID] == nil {
			c.sessionObjects[sessionID] = make(map[string]any)
		}
		c.sessionObjects[sessionID][name] = value
	}
}

func (c *Container) getInflightLocked(key string) *inflightBuild {
	if c.inflight == nil {
		return nil
	}
	return c.inflight[key]
}

func withResolverPath(ctx context.Context, name string) context.Context {
	path := resolverPathFromContext(ctx)
	path = append(path, name)
	return context.WithValue(ctx, resolverPathContextKey{}, path)
}

func resolverPathFromContext(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	path, _ := ctx.Value(resolverPathContextKey{}).([]string)
	return append([]string(nil), path...)
}

func hasCycle(ctx context.Context, name string) bool {
	path := resolverPathFromContext(ctx)
	for _, item := range path {
		if item == name {
			return true
		}
	}
	return false
}
