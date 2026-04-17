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

	if err := container.RegisterService("logger", func(ctx context.Context, resolver Resolver) (any, error) {
		buildCount++
		return "logger-instance", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	first, err := container.Resolve(context.Background(), "logger")
	if err != nil {
		t.Fatalf("resolve first failed: %v", err)
	}
	second, err := container.Resolve(context.Background(), "logger")
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

	if err := container.RegisterSession("order_component", func(ctx context.Context, resolver Resolver) (any, error) {
		buildCount++
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return "component:" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	ctx1 := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	ctx2 := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s2"))

	first, err := container.Resolve(ctx1, "order_component")
	if err != nil {
		t.Fatalf("resolve session1 failed: %v", err)
	}
	second, err := container.Resolve(ctx1, "order_component")
	if err != nil {
		t.Fatalf("resolve session1 second failed: %v", err)
	}
	third, err := container.Resolve(ctx2, "order_component")
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

	if err := container.RegisterService("config", func(ctx context.Context, resolver Resolver) (any, error) {
		return "cfg", nil
	}); err != nil {
		t.Fatalf("register config failed: %v", err)
	}
	if err := container.RegisterService("client", func(ctx context.Context, resolver Resolver) (any, error) {
		cfg, err := resolver.Resolve(ctx, "config")
		if err != nil {
			return nil, err
		}
		return "client:" + cfg.(string), nil
	}); err != nil {
		t.Fatalf("register client failed: %v", err)
	}

	value, err := container.Resolve(context.Background(), "client")
	if err != nil {
		t.Fatalf("resolve client failed: %v", err)
	}
	if value.(string) != "client:cfg" {
		t.Fatalf("unexpected client value: %v", value)
	}
}

func TestContainerResolveSessionRequiresContext(t *testing.T) {
	container := NewContainer()
	if err := container.RegisterSession("order_component", func(ctx context.Context, resolver Resolver) (any, error) {
		return "x", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	_, err := container.Resolve(context.Background(), "order_component")
	if !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("expected ErrSessionRequired, got %v", err)
	}
}

func TestContainerResolveCircularDependency(t *testing.T) {
	container := NewContainer()
	if err := container.RegisterService("a", func(ctx context.Context, resolver Resolver) (any, error) {
		_, err := resolver.Resolve(ctx, "b")
		return nil, err
	}); err != nil {
		t.Fatalf("register a failed: %v", err)
	}
	if err := container.RegisterService("b", func(ctx context.Context, resolver Resolver) (any, error) {
		_, err := resolver.Resolve(ctx, "a")
		return nil, err
	}); err != nil {
		t.Fatalf("register b failed: %v", err)
	}

	_, err := container.Resolve(context.Background(), "a")
	if !errors.Is(err, ErrCircularDependency) {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}

func TestContainerObjectManagement(t *testing.T) {
	container := NewContainer()
	if err := container.RegisterService("svc", func(ctx context.Context, resolver Resolver) (any, error) {
		return "svc-value", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}
	if err := container.RegisterSession("session_obj", func(ctx context.Context, resolver Resolver) (any, error) {
		return "session-value", nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	if _, err := container.Resolve(context.Background(), "svc"); err != nil {
		t.Fatalf("resolve service failed: %v", err)
	}
	if _, err := container.Resolve(ctx, "session_obj"); err != nil {
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

	container.DeleteService("svc")
	container.DeleteSessionObject("s1", "session_obj")
	if len(container.ServiceObjects()) != 0 {
		t.Fatalf("expected no service objects")
	}
	if len(container.SessionObjects("s1")) != 0 {
		t.Fatalf("expected no session objects")
	}
}
