package biz_component

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func TestKeyAccessors(t *testing.T) {
	serviceKey := ServiceKey[string]("svc")
	sessionKey := SessionKey[int]("sess")

	if serviceKey.Name() != "svc" || serviceKey.Scope() != ServiceScope {
		t.Fatalf("unexpected service key: %#v", serviceKey)
	}
	if sessionKey.Name() != "sess" || sessionKey.Scope() != SessionScope {
		t.Fatalf("unexpected session key: %#v", sessionKey)
	}
}

func TestRegisterValidation(t *testing.T) {
	key := ServiceKey[string]("svc")

	if err := Register[string](nil, key, func(context.Context, Resolver) (string, error) { return "", nil }); !errors.Is(err, ErrNilProvider) {
		t.Fatalf("expected ErrNilProvider for nil container, got %v", err)
	}
	container := NewContainer()
	if err := Register[string](container, key, nil); !errors.Is(err, ErrNilProvider) {
		t.Fatalf("expected ErrNilProvider for nil provider, got %v", err)
	}
	if err := RegisterService(container, SessionKey[string]("bad"), func(context.Context, Resolver) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected scope mismatch error")
	}
	if err := RegisterSession(container, ServiceKey[string]("bad"), func(context.Context, Resolver) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected scope mismatch error")
	}
	if err := container.RegisterAny("", ServiceScope, func(context.Context, Resolver) (any, error) { return nil, nil }); !errors.Is(err, ErrInvalidComponentName) {
		t.Fatalf("expected ErrInvalidComponentName, got %v", err)
	}
	if err := container.RegisterAny("svc", Scope("UNKNOWN"), func(context.Context, Resolver) (any, error) { return nil, nil }); err == nil {
		t.Fatal("expected unsupported scope error")
	}
}

func TestResolveAndObjectEdgeCases(t *testing.T) {
	key := ServiceKey[string]("svc")
	if _, err := Resolve[string](context.Background(), nil, key); !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}

	container := NewContainer()
	if err := container.RegisterAny("svc", ServiceScope, func(context.Context, Resolver) (any, error) {
		return 123, nil
	}); err != nil {
		t.Fatalf("register any failed: %v", err)
	}

	if _, err := Resolve[string](context.Background(), container, key); !errors.Is(err, ErrComponentTypeMismatch) {
		t.Fatalf("expected ErrComponentTypeMismatch, got %v", err)
	}
	if _, ok := ServiceObject[string](nil, key); ok {
		t.Fatal("expected nil container typed service lookup to fail")
	}
	if _, ok := SessionObject[string](nil, "s1", SessionKey[string]("sess")); ok {
		t.Fatal("expected nil container typed session lookup to fail")
	}
	if _, ok := ServiceObject[string](container, key); ok {
		t.Fatal("expected typed service lookup mismatch to fail")
	}
	if value, ok := container.ServiceObjectAny("svc"); !ok || value.(int) != 123 {
		t.Fatalf("unexpected service object any: %v %v", value, ok)
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

	if key, sessionID, err := container.resolveKey(context.Background(), "svc", ServiceScope); err != nil || key != "service:svc" || sessionID != "" {
		t.Fatalf("unexpected service resolve key: %q %q %v", key, sessionID, err)
	}
	if _, _, err := container.resolveKey(context.Background(), "sess", SessionScope); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("expected ErrSessionRequired, got %v", err)
	}

	ctx := withResolverPath(context.Background(), "a")
	ctx = withResolverPath(ctx, "b")
	if path := resolverPathFromContext(ctx); len(path) != 2 || path[0] != "a" || path[1] != "b" {
		t.Fatalf("unexpected resolver path: %#v", path)
	}
	if path := resolverPathFromContext(nil); path != nil {
		t.Fatalf("expected nil resolver path from nil ctx, got %#v", path)
	}
	if !hasCycle(ctx, "a") || hasCycle(ctx, "c") {
		t.Fatalf("unexpected cycle detection")
	}

	container.storeObjectLocked(ServiceScope, "", "svc", "value")
	container.storeObjectLocked(SessionScope, "s1", "sess", "session-value")
	if value, ok := container.cachedObjectLocked(ServiceScope, "", "svc"); !ok || value.(string) != "value" {
		t.Fatalf("unexpected cached service object: %v %v", value, ok)
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
