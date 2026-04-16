package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/ext_spi"
)

type testSPI interface {
	Handle(ctx context.Context, input string) (string, error)
}

type testSPIImpl struct {
	value string
}

func (i testSPIImpl) Handle(ctx context.Context, input string) (string, error) {
	return i.value + ":" + input, nil
}

func TestSPIContainer(t *testing.T) {
	template := ext_spi.NewTemplate(func(ctx context.Context, impl testSPI, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testSPI, input string) (string, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewSPIContainer[testSPI, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("audit", testSPIImpl{value: "impl-a"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.Register("audit", testSPIImpl{value: "impl-b"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	impls := container.Implementations("audit")
	if len(impls) != 2 {
		t.Fatalf("unexpected impls: %v", impls)
	}

	results, err := container.Execute(context.Background(), "audit", "input", ext_spi.All)
	if err != nil {
		t.Fatalf("expected execute success, got %v", err)
	}
	if len(results) != 2 || results[0] != "impl-a:input" || results[1] != "impl-b:input" {
		t.Fatalf("unexpected execute results: %v", results)
	}
}

func TestSPIContainerInvalidDefinition(t *testing.T) {
	template := ext_spi.NewTemplate(func(ctx context.Context, impl testSPI, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testSPI, input string) (string, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewSPIContainer[testSPI, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Register("", testSPIImpl{value: "impl-a"}); err == nil {
		t.Fatalf("expected invalid definition error")
	}
}

func TestSPIContainerNilTemplate(t *testing.T) {
	_, err := NewSPIContainer[testSPI, string, string](nil)
	if !errors.Is(err, ErrNilSPITemplate) {
		t.Fatalf("expected ErrNilSPITemplate, got %v", err)
	}
}
