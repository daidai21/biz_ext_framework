package service_manager

import (
	"context"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_component"
)

func TestComponentContainerResolveInSession(t *testing.T) {
	ctxContainer := NewCtxContainer()
	if _, err := ctxContainer.Create("s1"); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	container := NewComponentContainer(ctxContainer)
	if err := container.RegisterService("config", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		return "cfg", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}
	if err := container.RegisterSession("component", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		cfg, err := resolver.Resolve(ctx, "config")
		if err != nil {
			return nil, err
		}
		session, ok := ctxContainer.SessionFromContext(ctx)
		if !ok {
			t.Fatal("expected session in ctx")
		}
		return cfg.(string) + ":" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	value, err := container.ResolveInSession(context.Background(), "s1", "component")
	if err != nil {
		t.Fatalf("resolve in session failed: %v", err)
	}
	if value.(string) != "cfg:s1" {
		t.Fatalf("unexpected value: %v", value)
	}
}

func TestComponentContainerServiceObject(t *testing.T) {
	container := NewComponentContainer(nil)
	if err := container.RegisterService("logger", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		return "logger", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	if _, err := container.Resolve(context.Background(), "logger"); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	value, ok := container.ServiceObject("logger")
	if !ok || value.(string) != "logger" {
		t.Fatalf("unexpected service object: %v %v", value, ok)
	}
}
