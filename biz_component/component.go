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
	ErrInvalidComponentName      = errors.New("invalid component name")
	ErrInvalidComponentNamespace = errors.New("invalid component namespace")
	ErrNilProvider               = errors.New("nil component provider")
	ErrComponentNotFound         = errors.New("component not found")
	ErrSessionRequired           = errors.New("session is required")
	ErrCircularDependency        = errors.New("circular component dependency")
	ErrComponentTypeMismatch     = errors.New("component type mismatch")
	ErrNamespaceDependencyDenied = errors.New("component namespace dependency denied")
)

type Scope string

const (
	// GlobalScope is the recommended scope for one container-wide singleton.
	GlobalScope  Scope = "GLOBAL"
	SessionScope Scope = "SESSION"
)

type Namespace string

const (
	InfraNamespace      Namespace = "infra"
	RepositoryNamespace Namespace = "repository"
	ServiceNamespace    Namespace = "service"
	DomainNamespace     Namespace = "domain"
	CapabilityNamespace Namespace = "capability"
	BusinessNamespace   Namespace = "business"
	HandlerNamespace    Namespace = "handler"
)

type Key[T any] struct {
	name      string
	scope     Scope
	namespace Namespace
}

func GlobalKey[T any](name string) Key[T] {
	return GlobalKeyIn[T](HandlerNamespace, name)
}

func GlobalKeyIn[T any](namespace Namespace, name string) Key[T] {
	return Key[T]{
		name:      name,
		scope:     GlobalScope,
		namespace: namespace,
	}
}

func SessionKey[T any](name string) Key[T] {
	return SessionKeyIn[T](HandlerNamespace, name)
}

func SessionKeyIn[T any](namespace Namespace, name string) Key[T] {
	return Key[T]{
		name:      name,
		scope:     SessionScope,
		namespace: namespace,
	}
}

func (k Key[T]) Name() string {
	return k.name
}

func (k Key[T]) Scope() Scope {
	return k.scope
}

func (k Key[T]) Namespace() Namespace {
	return k.namespace
}

type Provider[T any] func(ctx context.Context, resolver Resolver) (T, error)

type Resolver interface {
	ResolveAny(ctx context.Context, name string) (any, error)
}

type definition struct {
	scope     Scope
	namespace Namespace
	provider  func(ctx context.Context, resolver Resolver) (any, error)
}

type inflightBuild struct {
	wait  chan struct{}
	value any
	err   error
}

type resolverPathContextKey struct{}

type resolverFrame struct {
	name      string
	namespace Namespace
}

// Container is a concurrency-safe IOC container supporting both global-level
// and session-level object management.
type Container struct {
	mu sync.Mutex

	definitions map[string]definition

	globalObjects  map[string]any
	sessionObjects map[string]map[string]any
	inflight       map[string]*inflightBuild
}

var _ Resolver = (*Container)(nil)

func NewContainer() *Container {
	return &Container{}
}

func Register[T any](container *Container, key Key[T], provider Provider[T]) error {
	if container == nil {
		return ErrNilProvider
	}
	if provider == nil {
		return ErrNilProvider
	}
	return container.register(key.name, key.scope, key.namespace, func(ctx context.Context, resolver Resolver) (any, error) {
		return provider(ctx, resolver)
	})
}

func RegisterGlobal[T any](container *Container, key Key[T], provider Provider[T]) error {
	if key.scope != GlobalScope {
		return fmt.Errorf("component %q is not a global scope key", key.name)
	}
	return Register(container, key, provider)
}

func RegisterSession[T any](container *Container, key Key[T], provider Provider[T]) error {
	if key.scope != SessionScope {
		return fmt.Errorf("component %q is not a session scope key", key.name)
	}
	return Register(container, key, provider)
}

func Resolve[T any](ctx context.Context, resolver Resolver, key Key[T]) (T, error) {
	var zero T
	if resolver == nil {
		return zero, fmt.Errorf("%w: %q", ErrComponentNotFound, key.name)
	}

	value, err := resolver.ResolveAny(ctx, key.name)
	if err != nil {
		return zero, err
	}
	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("%w: %q", ErrComponentTypeMismatch, key.name)
	}
	return typed, nil
}

func GlobalObject[T any](container *Container, key Key[T]) (T, bool) {
	var zero T
	if container == nil {
		return zero, false
	}
	value, ok := container.globalObjectAny(key.name)
	if !ok {
		return zero, false
	}
	typed, ok := value.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

func SessionObject[T any](container *Container, sessionID string, key Key[T]) (T, bool) {
	var zero T
	if container == nil {
		return zero, false
	}
	value, ok := container.sessionObjectAny(sessionID, key.name)
	if !ok {
		return zero, false
	}
	typed, ok := value.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

func (c *Container) RegisterAny(name string, scope Scope, provider func(ctx context.Context, resolver Resolver) (any, error)) error {
	return c.register(name, scope, HandlerNamespace, provider)
}

func (c *Container) RegisterAnyIn(name string, scope Scope, namespace Namespace, provider func(ctx context.Context, resolver Resolver) (any, error)) error {
	return c.register(name, scope, namespace, provider)
}

func (c *Container) ResolveAny(ctx context.Context, name string) (any, error) {
	if name == "" {
		return nil, ErrInvalidComponentName
	}

	c.mu.Lock()
	definition, ok := c.definitions[name]
	if !ok {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: %q", ErrComponentNotFound, name)
	}
	if err := validateDependency(ctx, definition.namespace, name); err != nil {
		c.mu.Unlock()
		return nil, err
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

	resolveCtx := withResolverFrame(ctx, name, definition.namespace)
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

func (c *Container) ServiceObjectAny(name string) (any, bool) {
	return c.serviceObjectAny(name)
}

func (c *Container) SessionObjectAny(sessionID, name string) (any, bool) {
	return c.sessionObjectAny(sessionID, name)
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

func (c *Container) register(name string, scope Scope, namespace Namespace, provider func(ctx context.Context, resolver Resolver) (any, error)) error {
	if name == "" {
		return ErrInvalidComponentName
	}
	if provider == nil {
		return ErrNilProvider
	}
	if scope != ServiceScope && scope != SessionScope {
		return fmt.Errorf("unsupported component scope: %q", scope)
	}
	if err := namespace.Validate(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.definitions == nil {
		c.definitions = make(map[string]definition)
	}
	c.definitions[name] = definition{
		scope:     scope,
		namespace: namespace,
		provider:  provider,
	}
	return nil
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

func (c *Container) serviceObjectAny(name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cachedObjectLocked(ServiceScope, "", name)
}

func (c *Container) sessionObjectAny(sessionID, name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cachedObjectLocked(SessionScope, sessionID, name)
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
	return withResolverFrame(ctx, name, HandlerNamespace)
}

func withResolverFrame(ctx context.Context, name string, namespace Namespace) context.Context {
	path := resolverFramesFromContext(ctx)
	path = append(path, resolverFrame{name: name, namespace: namespace})
	return context.WithValue(ctx, resolverPathContextKey{}, path)
}

func resolverPathFromContext(ctx context.Context) []string {
	frames := resolverFramesFromContext(ctx)
	path := make([]string, 0, len(frames))
	for _, frame := range frames {
		path = append(path, frame.name)
	}
	return path
}

func resolverFramesFromContext(ctx context.Context) []resolverFrame {
	if ctx == nil {
		return nil
	}
	path, _ := ctx.Value(resolverPathContextKey{}).([]resolverFrame)
	return append([]resolverFrame(nil), path...)
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

func (n Namespace) Validate() error {
	switch n {
	case InfraNamespace, RepositoryNamespace, ServiceNamespace, DomainNamespace, CapabilityNamespace, BusinessNamespace, HandlerNamespace:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidComponentNamespace, n)
	}
}

func validateDependency(ctx context.Context, targetNamespace Namespace, targetName string) error {
	parent, ok := currentResolverFrame(ctx)
	if !ok {
		return nil
	}
	if canDepend(parent.namespace, targetNamespace) {
		return nil
	}
	return fmt.Errorf("%w: %s -> %s (%s -> %s)", ErrNamespaceDependencyDenied, parent.name, targetName, parent.namespace, targetNamespace)
}

func currentResolverFrame(ctx context.Context) (resolverFrame, bool) {
	path := resolverFramesFromContext(ctx)
	if len(path) == 0 {
		return resolverFrame{}, false
	}
	return path[len(path)-1], true
}

func canDepend(from Namespace, to Namespace) bool {
	switch from {
	case InfraNamespace, RepositoryNamespace:
		return false
	case ServiceNamespace:
		return to == InfraNamespace || to == RepositoryNamespace
	case DomainNamespace:
		return to == ServiceNamespace || to == InfraNamespace || to == RepositoryNamespace
	case CapabilityNamespace, BusinessNamespace:
		return to == DomainNamespace || to == ServiceNamespace || to == RepositoryNamespace || to == InfraNamespace
	case HandlerNamespace:
		return true
	default:
		return false
	}
}
