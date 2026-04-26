package biz_component

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func TestKeyAccessors(t *testing.T) {
	globalKey := GlobalKey[string]("svc")
	anotherGlobalKey := GlobalKey[string]("global")
	sessionKey := SessionKey[int]("sess")
	explicitKey := GlobalKeyIn[string](ServiceNamespace, "service")

	if globalKey.Name() != "svc" || globalKey.Scope() != GlobalScope || globalKey.Namespace() != HandlerNamespace {
		t.Fatalf("unexpected global key: %#v", globalKey)
	}
	if anotherGlobalKey.Name() != "global" || anotherGlobalKey.Scope() != GlobalScope || anotherGlobalKey.Namespace() != HandlerNamespace {
		t.Fatalf("unexpected second global key: %#v", anotherGlobalKey)
	}
	if sessionKey.Name() != "sess" || sessionKey.Scope() != SessionScope || sessionKey.Namespace() != HandlerNamespace {
		t.Fatalf("unexpected session key: %#v", sessionKey)
	}
	if explicitKey.Namespace() != ServiceNamespace {
		t.Fatalf("unexpected explicit namespace key: %#v", explicitKey)
	}
}

func TestRegisterValidation(t *testing.T) {
	key := GlobalKey[string]("svc")

	if err := Register[string](nil, key, func(context.Context, Resolver) (string, error) { return "", nil }); !errors.Is(err, ErrNilProvider) {
		t.Fatalf("expected ErrNilProvider for nil container, got %v", err)
	}
	container := NewContainer()
	if err := Register[string](container, key, nil); !errors.Is(err, ErrNilProvider) {
		t.Fatalf("expected ErrNilProvider for nil provider, got %v", err)
	}
	if err := RegisterGlobal(container, SessionKey[string]("bad"), func(context.Context, Resolver) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected scope mismatch error")
	}
	if err := RegisterSession(container, GlobalKey[string]("bad"), func(context.Context, Resolver) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected scope mismatch error")
	}
	if err := container.RegisterAny("", GlobalScope, func(context.Context, Resolver) (any, error) { return nil, nil }); !errors.Is(err, ErrInvalidComponentName) {
		t.Fatalf("expected ErrInvalidComponentName, got %v", err)
	}
	if err := container.RegisterAny("global", GlobalScope, func(context.Context, Resolver) (any, error) { return nil, nil }); err != nil {
		t.Fatalf("expected global scope registration success, got %v", err)
	}
	if err := container.RegisterAny("svc", Scope("UNKNOWN"), func(context.Context, Resolver) (any, error) { return nil, nil }); err == nil {
		t.Fatal("expected unsupported scope error")
	}
	if err := container.RegisterAnyIn("svc", GlobalScope, Namespace("unknown"), func(context.Context, Resolver) (any, error) { return nil, nil }); !errors.Is(err, ErrInvalidComponentNamespace) {
		t.Fatalf("expected ErrInvalidComponentNamespace, got %v", err)
	}
}

func TestResolveAndObjectEdgeCases(t *testing.T) {
	key := GlobalKey[string]("svc")
	if _, err := Resolve[string](context.Background(), nil, key); !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}

	container := NewContainer()
	if err := container.RegisterAny("svc", GlobalScope, func(context.Context, Resolver) (any, error) {
		return 123, nil
	}); err != nil {
		t.Fatalf("register any failed: %v", err)
	}

	if _, err := Resolve[string](context.Background(), container, key); !errors.Is(err, ErrComponentTypeMismatch) {
		t.Fatalf("expected ErrComponentTypeMismatch, got %v", err)
	}
	if _, ok := GlobalObject[string](nil, GlobalKey[string]("svc")); ok {
		t.Fatal("expected nil container typed global lookup to fail")
	}
	if _, ok := SessionObject[string](nil, "s1", SessionKey[string]("sess")); ok {
		t.Fatal("expected nil container typed session lookup to fail")
	}
	if value, ok := container.GlobalObjectAny("svc"); !ok || value.(int) != 123 {
		t.Fatalf("unexpected global object any: %v %v", value, ok)
	}
	if _, err := container.ResolveAny(context.Background(), ""); !errors.Is(err, ErrInvalidComponentName) {
		t.Fatalf("expected ErrInvalidComponentName, got %v", err)
	}
	if _, err := container.ResolveAny(context.Background(), "missing"); !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}
}

func TestSessionObjectAnyAndClearSession(t *testing.T) {
	container := NewContainer()
	key := SessionKey[string]("component")
	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))

	if err := RegisterSession(container, key, func(context.Context, Resolver) (string, error) {
		return "value", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	if _, err := Resolve(ctx, container, key); err != nil {
		t.Fatalf("resolve session failed: %v", err)
	}
	if value, ok := container.SessionObjectAny("s1", "component"); !ok || value.(string) != "value" {
		t.Fatalf("unexpected session object any: %v %v", value, ok)
	}
	container.ClearSession("s1")
	if _, ok := container.SessionObjectAny("s1", "component"); ok {
		t.Fatal("expected cleared session object to be absent")
	}
}

func TestPrivateHelpers(t *testing.T) {
	container := NewContainer()

	if key, sessionID, err := container.resolveKey(context.Background(), "global", GlobalScope); err != nil || key != "global:global" || sessionID != "" {
		t.Fatalf("unexpected global resolve key: %q %q %v", key, sessionID, err)
	}
	if _, _, err := container.resolveKey(context.Background(), "sess", SessionScope); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("expected ErrSessionRequired, got %v", err)
	}

	ctx := withResolverPath(context.Background(), "a")
	ctx = withResolverFrame(ctx, "b", SessionScope, ServiceNamespace)
	if path := resolverPathFromContext(ctx); len(path) != 2 || path[0] != "GLOBAL:a" || path[1] != "SESSION:b" {
		t.Fatalf("unexpected resolver path: %#v", path)
	}
	if frames := resolverFramesFromContext(ctx); len(frames) != 2 || frames[1].namespace != ServiceNamespace || frames[1].scope != SessionScope {
		t.Fatalf("unexpected resolver frames: %#v", frames)
	}
	if path := resolverPathFromContext(nil); len(path) != 0 {
		t.Fatalf("expected empty resolver path from nil ctx, got %#v", path)
	}
	if !hasCycle(ctx, "a", GlobalScope) || hasCycle(ctx, "a", SessionScope) || hasCycle(ctx, "c", GlobalScope) {
		t.Fatalf("unexpected cycle detection")
	}
	if current, ok := currentResolverFrame(ctx); !ok || current.name != "b" || current.scope != SessionScope || current.namespace != ServiceNamespace {
		t.Fatalf("unexpected current resolver frame: %#v %v", current, ok)
	}
	if _, ok := currentResolverFrame(context.Background()); ok {
		t.Fatal("expected no current resolver frame in empty ctx")
	}

	container.storeObjectLocked(GlobalScope, "", "svc", "value")
	container.storeObjectLocked(SessionScope, "s1", "sess", "session-value")
	if value, ok := container.cachedObjectLocked(GlobalScope, "", "svc"); !ok || value.(string) != "value" {
		t.Fatalf("unexpected cached global object: %v %v", value, ok)
	}
	if value, ok := container.cachedObjectLocked(SessionScope, "s1", "sess"); !ok || value.(string) != "session-value" {
		t.Fatalf("unexpected cached session object: %v %v", value, ok)
	}
	if _, ok := container.cachedObjectLocked(Scope("UNKNOWN"), "", "svc"); ok {
		t.Fatal("expected unknown scope cache lookup to miss")
	}
	if inflight := container.getInflightLocked("missing"); inflight != nil {
		t.Fatalf("expected nil inflight, got %#v", inflight)
	}
}

func TestNamespaceValidationAndDependencyRules(t *testing.T) {
	valid := []Namespace{
		InfraNamespace,
		RepositoryNamespace,
		ServiceNamespace,
		DomainNamespace,
		CapabilityNamespace,
		BusinessNamespace,
		HandlerNamespace,
	}
	for _, namespace := range valid {
		if err := namespace.Validate(); err != nil {
			t.Fatalf("expected namespace %q valid, got %v", namespace, err)
		}
	}
	if err := Namespace("bad").Validate(); !errors.Is(err, ErrInvalidComponentNamespace) {
		t.Fatalf("expected ErrInvalidComponentNamespace, got %v", err)
	}

	if canDepend(ServiceNamespace, RepositoryNamespace) != true {
		t.Fatal("expected service -> repository allowed")
	}
	if canDepend(ServiceNamespace, DomainNamespace) {
		t.Fatal("expected service -> domain denied")
	}
	if canDepend(InfraNamespace, RepositoryNamespace) {
		t.Fatal("expected infra -> repository denied")
	}
	if canDepend(HandlerNamespace, BusinessNamespace) != true {
		t.Fatal("expected handler -> business allowed")
	}
	if canDepend(BusinessNamespace, CapabilityNamespace) {
		t.Fatal("expected business -> capability denied")
	}
}

func TestNamespaceDependencyEnforced(t *testing.T) {
	container := NewContainer()
	serviceKey := GlobalKeyIn[string](ServiceNamespace, "service")
	domainKey := GlobalKeyIn[string](DomainNamespace, "domain")

	if err := Register(container, domainKey, func(context.Context, Resolver) (string, error) {
		return "domain", nil
	}); err != nil {
		t.Fatalf("register domain failed: %v", err)
	}
	if err := Register(container, serviceKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return Resolve(ctx, resolver, domainKey)
	}); err != nil {
		t.Fatalf("register global failed: %v", err)
	}

	_, err := Resolve(context.Background(), container, serviceKey)
	if !errors.Is(err, ErrNamespaceDependencyDenied) {
		t.Fatalf("expected ErrNamespaceDependencyDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "service -> domain") {
		t.Fatalf("expected namespace path in error, got %v", err)
	}
}

func TestNamespaceDependencyAllowed(t *testing.T) {
	container := NewContainer()
	repoKey := GlobalKeyIn[string](RepositoryNamespace, "repo")
	serviceKey := GlobalKeyIn[string](ServiceNamespace, "service")
	domainKey := GlobalKeyIn[string](DomainNamespace, "domain")
	capabilityKey := GlobalKeyIn[string](CapabilityNamespace, "capability")
	businessKey := GlobalKeyIn[string](BusinessNamespace, "business")
	handlerKey := GlobalKeyIn[string](HandlerNamespace, "handler")

	for _, item := range []struct {
		key      Key[string]
		provider Provider[string]
	}{
		{repoKey, func(context.Context, Resolver) (string, error) { return "repo", nil }},
		{serviceKey, func(ctx context.Context, resolver Resolver) (string, error) {
			repo, err := Resolve(ctx, resolver, repoKey)
			if err != nil {
				return "", err
			}
			return "service:" + repo, nil
		}},
		{domainKey, func(ctx context.Context, resolver Resolver) (string, error) {
			svc, err := Resolve(ctx, resolver, serviceKey)
			if err != nil {
				return "", err
			}
			return "domain:" + svc, nil
		}},
		{capabilityKey, func(ctx context.Context, resolver Resolver) (string, error) {
			domain, err := Resolve(ctx, resolver, domainKey)
			if err != nil {
				return "", err
			}
			return "capability:" + domain, nil
		}},
		{businessKey, func(ctx context.Context, resolver Resolver) (string, error) {
			domain, err := Resolve(ctx, resolver, domainKey)
			if err != nil {
				return "", err
			}
			return "business:" + domain, nil
		}},
		{handlerKey, func(ctx context.Context, resolver Resolver) (string, error) {
			business, err := Resolve(ctx, resolver, businessKey)
			if err != nil {
				return "", err
			}
			return "handler:" + business, nil
		}},
	} {
		if err := Register(container, item.key, item.provider); err != nil {
			t.Fatalf("register %q failed: %v", item.key.Name(), err)
		}
	}

	value, err := Resolve(context.Background(), container, handlerKey)
	if err != nil {
		t.Fatalf("resolve handler failed: %v", err)
	}
	if value != "handler:business:domain:service:repo" {
		t.Fatalf("unexpected value: %q", value)
	}
}
