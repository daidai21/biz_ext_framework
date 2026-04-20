package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_process"
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

func TestProcessContainerRun(t *testing.T) {
	var order []string
	process := biz_process.Process{
		Layers: []biz_process.ProcessLayer{
			{
				Name: "prepare",
				Nodes: []biz_process.ProcessNode{
					biz_process.TaskProcessNode{Name: "prepare", Task: func(ctx context.Context) error {
						order = append(order, "prepare")
						return nil
					}},
				},
			},
			{
				Name: "finalize",
				Nodes: []biz_process.ProcessNode{
					biz_process.TaskProcessNode{Name: "finalize", Task: func(ctx context.Context) error {
						order = append(order, "finalize")
						return nil
					}},
				},
			},
		},
	}

	container := NewProcessContainer()
	if err := container.Register("order_flow", process); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Run(context.Background(), "order_flow"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(order) != 2 || order[0] != "prepare" || order[1] != "finalize" {
		t.Fatalf("unexpected process order: %v", order)
	}
}

func TestProcessContainerRunNotFound(t *testing.T) {
	container := NewProcessContainer()
	err := container.Run(context.Background(), "missing")
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
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

func TestExtProcessContainerRegisterWithAction(t *testing.T) {
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
	if err := container.RegisterWithAction("audit", testExtProcessImpl{value: "impl-skip"}, ext_process.Skip); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	impls := container.Implementations("audit")
	if len(impls) != 1 {
		t.Fatalf("expected skip to keep existing implementations, got %d", len(impls))
	}

	if err := container.RegisterWithAction("audit", testExtProcessImpl{value: "impl-overwrite", stop: true}, ext_process.Overwrite); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	results, err := container.Execute(context.Background(), "audit", "input", ext_process.Serial)
	if err != nil {
		t.Fatalf("expected execute success, got %v", err)
	}
	if len(results) != 1 || results[0] != "impl-overwrite:input" {
		t.Fatalf("unexpected overwrite results: %v", results)
	}
}

func TestExtProcessContainerRegisterWithAppendType(t *testing.T) {
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
	if err := container.RegisterWithAppendType("audit", testExtProcessImpl{value: "impl-before"}, ext_process.AppendBefore); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.RegisterWithAppendType("audit", testExtProcessImpl{value: "impl-parallel"}, ext_process.AppendParallel); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	results, err := container.Execute(context.Background(), "audit", "input", ext_process.Serial)
	if err != nil {
		t.Fatalf("expected execute success, got %v", err)
	}
	if len(results) != 3 ||
		results[0] != "impl-before:input" ||
		results[1] != "impl-a:input" ||
		results[2] != "impl-parallel:input" {
		t.Fatalf("unexpected append type results: %v", results)
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

func TestExtProcessContainerInvalidAction(t *testing.T) {
	template := ext_process.NewTemplate(func(ctx context.Context, impl testExtProcess, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testExtProcess, input string) (string, bool, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewExtProcessContainer[testExtProcess, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.RegisterWithAction("audit", testExtProcessImpl{value: "impl-a"}, ext_process.DefinitionAction("UNKNOWN")); err == nil {
		t.Fatalf("expected invalid action error")
	}
}

func TestExtProcessContainerInvalidAppendType(t *testing.T) {
	template := ext_process.NewTemplate(func(ctx context.Context, impl testExtProcess, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testExtProcess, input string) (string, bool, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewExtProcessContainer[testExtProcess, string, string](template)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.RegisterWithAppendType("audit", testExtProcessImpl{value: "impl-a"}, ext_process.AppendType("UNKNOWN")); err == nil {
		t.Fatalf("expected invalid append type error")
	}
}

func TestExtProcessContainerNilTemplate(t *testing.T) {
	_, err := NewExtProcessContainer[testExtProcess, string, string](nil)
	if !errors.Is(err, ErrNilExtProcessTemplate) {
		t.Fatalf("expected ErrNilExtProcessTemplate, got %v", err)
	}
}
