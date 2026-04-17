package biz_component

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func TestContainerResolveServiceSingleton(t *testing.T) {
	container := NewContainer()
	buildCount := 0
	loggerKey := ServiceKey[string]("logger")

	if err := RegisterService(container, loggerKey, func(ctx context.Context, resolver Resolver) (string, error) {
		buildCount++
		return "logger-instance", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
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
		t.Fatalf("expected singleton service object")
	}
	if buildCount != 1 {
		t.Fatalf("expected build count 1, got %d", buildCount)
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

func TestContainerResolveServiceDependency(t *testing.T) {
	container := NewContainer()
	configKey := ServiceKey[string]("config")
	clientKey := ServiceKey[string]("client")

	if err := RegisterService(container, configKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "cfg", nil
	}); err != nil {
		t.Fatalf("register config failed: %v", err)
	}
	if err := RegisterService(container, clientKey, func(ctx context.Context, resolver Resolver) (string, error) {
		cfg, err := Resolve(ctx, resolver, configKey)
		if err != nil {
			return "", err
		}
		return "client:" + cfg, nil
	}); err != nil {
		t.Fatalf("register client failed: %v", err)
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
	aKey := ServiceKey[string]("a")
	bKey := ServiceKey[string]("b")

	if err := RegisterService(container, aKey, func(ctx context.Context, resolver Resolver) (string, error) {
		_, err := Resolve(ctx, resolver, bKey)
		return "", err
	}); err != nil {
		t.Fatalf("register a failed: %v", err)
	}
	if err := RegisterService(container, bKey, func(ctx context.Context, resolver Resolver) (string, error) {
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
	serviceKey := ServiceKey[string]("svc")
	sessionKey := SessionKey[string]("session_obj")

	if err := RegisterService(container, serviceKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "svc-value", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}
	if err := RegisterSession(container, sessionKey, func(ctx context.Context, resolver Resolver) (string, error) {
		return "session-value", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	if _, err := Resolve(context.Background(), container, serviceKey); err != nil {
		t.Fatalf("resolve service failed: %v", err)
	}
	if _, err := Resolve(ctx, container, sessionKey); err != nil {
		t.Fatalf("resolve session failed: %v", err)
	}

	if len(container.ServiceObjects()) != 1 {
		t.Fatalf("expected 1 service object")
	}
	if len(container.SessionObjects("s1")) != 1 {
		t.Fatalf("expected 1 session object")
	}
	if len(container.ServiceNames()) != 1 || container.ServiceNames()[0] != "svc" {
		t.Fatalf("unexpected service names: %v", container.ServiceNames())
	}
	if len(container.SessionNames("s1")) != 1 || container.SessionNames("s1")[0] != "session_obj" {
		t.Fatalf("unexpected session names: %v", container.SessionNames("s1"))
	}

	serviceValue, ok := ServiceObject(container, serviceKey)
	if !ok || serviceValue != "svc-value" {
		t.Fatalf("unexpected typed service object: %v %v", serviceValue, ok)
	}
	sessionValue, ok := SessionObject(container, "s1", sessionKey)
	if !ok || sessionValue != "session-value" {
		t.Fatalf("unexpected typed session object: %v %v", sessionValue, ok)
	}

	container.DeleteService("svc")
	container.DeleteSessionObject("s1", "session_obj")
	if len(container.ServiceObjects()) != 0 {
		t.Fatalf("expected no service objects")
	}
	if len(container.SessionObjects("s1")) != 0 {
		t.Fatalf("expected no session objects")
	}
}
