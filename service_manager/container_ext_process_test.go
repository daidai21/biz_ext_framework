package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/ext_process"
)

type testExtProcess interface {
	Handle(ctx context.Context, input string) (string, bool, error)
}

type testExtProcessImpl struct {
	value string
	stop  bool
}

func (i testExtProcessImpl) Handle(ctx context.Context, input string) (string, bool, error) {
	return i.value + ":" + input, !i.stop, nil
}

func TestExtProcessContainer(t *testing.T) {
	template := ext_process.NewTemplate(func(ctx context.Context, impl testExtProcess, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testExtProcess, input string) (string, bool, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewExtProcessContainer[testExtProcess, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("audit", testExtProcessImpl{value: "impl-a"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.Register("audit", testExtProcessImpl{value: "impl-b", stop: true}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.Register("audit", testExtProcessImpl{value: "impl-c"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	results, err := container.Execute(context.Background(), "audit", "input", ext_process.Serial)
	if err != nil {
		t.Fatalf("expected execute success, got %v", err)
	}
	if len(results) != 2 || results[0] != "impl-a:input" || results[1] != "impl-b:input" {
		t.Fatalf("unexpected results: %v", results)
	}
}

func TestExtProcessContainerInvalidDefinition(t *testing.T) {
	template := ext_process.NewTemplate(func(ctx context.Context, impl testExtProcess, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testExtProcess, input string) (string, bool, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewExtProcessContainer[testExtProcess, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("", testExtProcessImpl{value: "impl-a"}); err == nil {
		t.Fatalf("expected invalid definition error")
	}
}

func TestExtProcessContainerNilTemplate(t *testing.T) {
	_, err := NewExtProcessContainer[testExtProcess, string, string](nil)
	if !errors.Is(err, ErrNilExtProcessTemplate) {
		t.Fatalf("expected ErrNilExtProcessTemplate, got %v", err)
	}
}
