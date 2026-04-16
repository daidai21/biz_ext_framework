package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/ext_interceptor"
)

type testInterceptorSPI interface {
	Handle(ctx context.Context, input string, next ext_interceptor.Handler[string, string]) (string, error)
}

type testInterceptorImpl struct {
	name string
}

func (i testInterceptorImpl) Handle(ctx context.Context, input string, next ext_interceptor.Handler[string, string]) (string, error) {
	output, err := next(ctx, input+"->"+i.name)
	if err != nil {
		return "", err
	}
	return output + "|after:" + i.name, nil
}

func TestInterceptorContainer(t *testing.T) {
	template := ext_interceptor.NewTemplate(func(ctx context.Context, impl testInterceptorSPI, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testInterceptorSPI, input string, next ext_interceptor.Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})
	container, err := NewInterceptorContainer[testInterceptorSPI, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("rpc", testInterceptorImpl{name: "a"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.Register("rpc", testInterceptorImpl{name: "b"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(container.Interceptors("rpc")) != 2 {
		t.Fatalf("expected 2 interceptors")
	}

	output, err := container.Execute(context.Background(), "rpc", "start", func(ctx context.Context, input string) (string, error) {
		return input + "->final", nil
	})
	if err != nil {
		t.Fatalf("expected execute success, got %v", err)
	}
	if output != "start->a->b->final|after:b|after:a" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestInterceptorContainerInvalidDefinition(t *testing.T) {
	template := ext_interceptor.NewTemplate(func(ctx context.Context, impl testInterceptorSPI, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testInterceptorSPI, input string, next ext_interceptor.Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})
	container, err := NewInterceptorContainer[testInterceptorSPI, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("", testInterceptorImpl{name: "a"}); err == nil {
		t.Fatalf("expected invalid definition error")
	}
}

func TestInterceptorContainerNilTemplate(t *testing.T) {
	_, err := NewInterceptorContainer[testInterceptorSPI, string, string](nil)
	if !errors.Is(err, ErrNilInterceptorTemplate) {
		t.Fatalf("expected ErrNilInterceptorTemplate, got %v", err)
	}
}
