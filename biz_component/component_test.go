package biz_component

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func TestContainerResolveGlobalSingleton(t *testing.T) {
	container := NewContainer()
	buildCount := 0
	loggerKey := GlobalKey[string]("logger")

	if err := RegisterGlobal(container, loggerKey, func(ctx context.Context, resolver Resolver) (string, error) {
		buildCount++
		return "logger-instance", nil
	}); err != nil {
		t.Fatalf("register global failed: %v", err)
	}

	first, err := Resolve(context.Background(), container, loggerKey)
	if err != nil {
		t.Fatalf("resolve first failed: %v", err)
	}
	second, err := Resolve(context.Background(), container, loggerKey)
	if err != nil {
		t.Fatalf("resolve second failed: %v", err)
	}

	if first != second {
		t.Fatalf("expected singleton global object")
	}
	if buildCount != 1 {
		t.Fatalf("expected build count 1, got %d", buildCount)
	}
	if value, ok := GlobalObject(container, loggerKey); !ok || value != "logger-instance" {
		t.Fatalf("unexpected typed global object: %v %v", value, ok)
	}
}

func TestContainerResolveSessionScoped(t *testing.T) {
	container := NewContainer()
	buildCount := 0
	componentKey := SessionKey[string]("order_component")

	if err := RegisterSession(container, componentKey, func(ctx context.Context, resolver Resolver) (string, error) {
		buildCount++
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return "component:" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	ctx1 := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	ctx2 := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s2"))

	first, err := Resolve(ctx1, container, componentKey)
	if err != nil {
		t.Fatalf("resolve session1 failed: %v", err)
	}
	second, err := Resolve(ctx1, container, componentKey)
	if err != nil {
		t.Fatalf("resolve session1 second failed: %v", err)
	}
	third, err := Resolve(ctx2, container, componentKey)
	if err != nil {
		t.Fatalf("resolve session2 failed: %v", err)
	}

	if first != second {
		t.Fatalf("expected same object in same session")
	}
	if first == third {
		t.Fatalf("expected different object across sessions")
	}
	if buildCount != 2 {
		t.Fatalf("expected build count 2, got %d", buildCount)
	}
}

func TestContainerResolveSameNameAcrossScopes(t *testing.T) {
	container := NewContainer()
	componentGlobalKey := GlobalKey[string]("component")
	componentSessionKey := SessionKey[string]("component")

	if err := RegisterGlobal(container, componentGlobalKey, func(context.Context, Resolver) (string, error) {
		return "global-component", nil
	}); err != nil {
		t.Fatalf("register global failed: %v", err)
	}
	if err := RegisterSession(container, componentSessionKey, func(ctx context.Context, resolver Resolver) (string, error) {
		globalValue, err := Resolve(ctx, resolver, componentGlobalKey)
		if err != nil {
			return "", err
		}
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return globalValue + ":" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	sessionCtx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))

	globalValue, err := Resolve(context.Background(), container, componentGlobalKey)
	if err != nil {
		t.Fatalf("resolve global failed: %v", err)
	}
	if globalValue != "global-component" {
		t.Fatalf("unexpected global value: %v", globalValue)
	}

	globalInSession, err := Resolve(sessionCtx, container, componentGlobalKey)
	if err != nil {
		t.Fatalf("resolve global in session failed: %v", err)
	}
	if globalInSession != "global-component" {
		t.Fatalf("unexpected global in session value: %v", globalInSession)
	}

	sessionValue, err := Resolve(sessionCtx, container, componentSessionKey)
	if err != nil {
		t.Fatalf("resolve session failed: %v", err)
	}
	if sessionValue != "global-component:s1" {
		t.Fatalf("unexpected session value: %v", sessionValue)
	}

	value, err := container.ResolveAny(context.Background(), "component")
	if err != nil {
		t.Fatalf("resolve any global failed: %v", err)
	}
	if value.(string) != "global-component" {
		t.Fatalf("unexpected resolve any global value: %v", value)
	}

	value, err = container.ResolveAny(sessionCtx, "component")
	if err != nil {
		t.Fatalf("resolve any session failed: %v", err)
	}
	if value.(string) != "global-component:s1" {
		t.Fatalf("unexpected resolve any session value: %v", value)
	}
}

func TestContainerResolveGlobalDependency(t *testing.T) {
	container := NewContainer()
	configKey := GlobalKey[string]("config")
	clientKey := GlobalKey[string]("client")

	if err := RegisterGlobal(container, configKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "cfg", nil
	}); err != nil {
		t.Fatalf("register global config failed: %v", err)
	}
	if err := RegisterGlobal(container, clientKey, func(ctx context.Context, resolver Resolver) (string, error) {
		cfg, err := Resolve(ctx, resolver, configKey)
		if err != nil {
			return "", err
		}
		return "client:" + cfg, nil
	}); err != nil {
		t.Fatalf("register global client failed: %v", err)
	}

	value, err := Resolve(context.Background(), container, clientKey)
	if err != nil {
		t.Fatalf("resolve client failed: %v", err)
	}
	if value != "client:cfg" {
		t.Fatalf("unexpected client value: %v", value)
	}
}

func TestContainerResolveSessionRequiresContext(t *testing.T) {
	container := NewContainer()
	componentKey := SessionKey[string]("order_component")

	if err := RegisterSession(container, componentKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "x", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	_, err := Resolve(context.Background(), container, componentKey)
	if !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("expected ErrSessionRequired, got %v", err)
	}
}

func TestContainerResolveCircularDependency(t *testing.T) {
	container := NewContainer()
	aKey := GlobalKey[string]("a")
	bKey := GlobalKey[string]("b")

	if err := RegisterGlobal(container, aKey, func(ctx context.Context, resolver Resolver) (string, error) {
		_, err := Resolve(ctx, resolver, bKey)
		return "", err
	}); err != nil {
		t.Fatalf("register a failed: %v", err)
	}
	if err := RegisterGlobal(container, bKey, func(ctx context.Context, resolver Resolver) (string, error) {
		_, err := Resolve(ctx, resolver, aKey)
		return "", err
	}); err != nil {
		t.Fatalf("register b failed: %v", err)
	}

	_, err := Resolve(context.Background(), container, aKey)
	if !errors.Is(err, ErrCircularDependency) {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}

func TestContainerObjectManagement(t *testing.T) {
	container := NewContainer()
	globalKey := GlobalKey[string]("svc")
	sessionKey := SessionKey[string]("session_obj")

	if err := RegisterGlobal(container, globalKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "svc-value", nil
	}); err != nil {
		t.Fatalf("register global failed: %v", err)
	}
	if err := RegisterSession(container, sessionKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "session-value", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	if _, err := Resolve(context.Background(), container, globalKey); err != nil {
		t.Fatalf("resolve global failed: %v", err)
	}
	if _, err := Resolve(ctx, container, sessionKey); err != nil {
		t.Fatalf("resolve session failed: %v", err)
	}

	if len(container.GlobalObjects()) != 1 {
		t.Fatalf("expected 1 global object")
	}
	if len(container.SessionObjects("s1")) != 1 {
		t.Fatalf("expected 1 session object")
	}
	if len(container.GlobalNames()) != 1 || container.GlobalNames()[0] != "svc" {
		t.Fatalf("unexpected global names: %v", container.GlobalNames())
	}
	if len(container.SessionNames("s1")) != 1 || container.SessionNames("s1")[0] != "session_obj" {
		t.Fatalf("unexpected session names: %v", container.SessionNames("s1"))
	}

	globalValue, ok := GlobalObject(container, globalKey)
	if !ok || globalValue != "svc-value" {
		t.Fatalf("unexpected typed global object: %v %v", globalValue, ok)
	}
	sessionValue, ok := SessionObject(container, "s1", sessionKey)
	if !ok || sessionValue != "session-value" {
		t.Fatalf("unexpected typed session object: %v %v", sessionValue, ok)
	}

	container.DeleteGlobal("svc")
	container.DeleteSessionObject("s1", "session_obj")
	if len(container.GlobalObjects()) != 0 {
		t.Fatalf("expected no global objects")
	}
	if len(container.SessionObjects("s1")) != 0 {
		t.Fatalf("expected no session objects")
	}
}
