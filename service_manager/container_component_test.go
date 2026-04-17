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
	configKey := biz_component.ServiceKey[string]("config")
	componentKey := biz_component.SessionKey[string]("component")

	if err := biz_component.RegisterService(container.Container(), configKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		return "cfg", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}
	if err := biz_component.RegisterSession(container.Container(), componentKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		cfg, err := biz_component.Resolve(ctx, resolver, configKey)
		if err != nil {
			return "", err
		}
		session, ok := ctxContainer.SessionFromContext(ctx)
		if !ok {
			t.Fatal("expected session in ctx")
		}
		return cfg + ":" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	value, err := container.ResolveAnyInSession(context.Background(), "s1", componentKey.Name())
	if err != nil {
		t.Fatalf("resolve in session failed: %v", err)
	}
	if value.(string) != "cfg:s1" {
		t.Fatalf("unexpected value: %v", value)
	}
}

func TestComponentContainerServiceObject(t *testing.T) {
	container := NewComponentContainer(nil)
	loggerKey := biz_component.ServiceKey[string]("logger")

	if err := biz_component.RegisterService(container.Container(), loggerKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		return "logger", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	if _, err := biz_component.Resolve(context.Background(), container.Container(), loggerKey); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	value, ok := biz_component.ServiceObject(container.Container(), loggerKey)
	if !ok || value != "logger" {
		t.Fatalf("unexpected service object: %v %v", value, ok)
	}
}
