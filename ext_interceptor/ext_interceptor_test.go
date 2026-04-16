package ext_interceptor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type testInterceptor interface {
	Name() string
	Handle(ctx context.Context, input string, next Handler[string, string]) (string, error)
}

type testInterceptorImpl struct {
	name string
}

func (i testInterceptorImpl) Name() string {
	return i.name
}

func (i testInterceptorImpl) Handle(ctx context.Context, input string, next Handler[string, string]) (string, error) {
	output, err := next(ctx, input+"->"+i.name)
	if err != nil {
		return "", err
	}
	return output + "|after:" + i.name, nil
}

func TestNewTemplate(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testInterceptor, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testInterceptor, input string, next Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})

	output, err := template(context.Background(), []testInterceptor{
		testInterceptorImpl{name: "a"},
		testInterceptorImpl{name: "b"},
	}, "start", func(ctx context.Context, input string) (string, error) {
		return input + "->final", nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if output != "start->a->b->final|after:b|after:a" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestNewTemplateMatch(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testInterceptor, input string) (bool, error) {
		return impl.Name() != "skip", nil
	}, func(ctx context.Context, impl testInterceptor, input string, next Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})

	output, err := template(context.Background(), []testInterceptor{
		testInterceptorImpl{name: "skip"},
		testInterceptorImpl{name: "run"},
	}, "start", func(ctx context.Context, input string) (string, error) {
		return input, nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if strings.Contains(output, "skip") || !strings.Contains(output, "run") {
		t.Fatalf("unexpected matched output: %s", output)
	}
}

func TestNewTemplateFinalRequired(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testInterceptor, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testInterceptor, input string, next Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})

	_, err := template(context.Background(), nil, "start", nil)
	if err == nil || !strings.Contains(err.Error(), "final handler is required") {
		t.Fatalf("expected final handler required error, got %v", err)
	}
}

func TestNewTemplateInterceptorRequired(t *testing.T) {
	template := NewTemplate[testInterceptor, string, string](nil, nil)
	_, err := template(context.Background(), nil, "start", func(ctx context.Context, input string) (string, error) {
		return input, nil
	})
	if err == nil || !strings.Contains(err.Error(), "interceptor func is required") {
		t.Fatalf("expected interceptor func required error, got %v", err)
	}
}

func TestNewTemplateMatchError(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testInterceptor, input string) (bool, error) {
		return false, errors.New("match failed")
	}, func(ctx context.Context, impl testInterceptor, input string, next Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})

	_, err := template(context.Background(), []testInterceptor{testInterceptorImpl{name: "a"}}, "start", func(ctx context.Context, input string) (string, error) {
		return input, nil
	})
	if err == nil || !strings.Contains(err.Error(), "match failed") {
		t.Fatalf("expected match failed error, got %v", err)
	}
}
