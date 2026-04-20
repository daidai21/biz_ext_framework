package ext_process

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type testInput struct {
	scene string
}

type testProc interface {
	Name() string
	Match(ctx context.Context, input testInput) (bool, error)
	Handle(ctx context.Context, input testInput) (string, bool, error)
}

type primaryProc struct{}

func (primaryProc) Name() string { return "primary" }
func (primaryProc) Match(_ context.Context, input testInput) (bool, error) {
	return input.scene == "ORDER", nil
}
func (primaryProc) Handle(_ context.Context, _ testInput) (string, bool, error) {
	return "primary", true, nil
}

type secondaryProc struct{}

func (secondaryProc) Name() string { return "secondary" }
func (secondaryProc) Match(_ context.Context, input testInput) (bool, error) {
	return input.scene == "ORDER", nil
}
func (secondaryProc) Handle(_ context.Context, _ testInput) (string, bool, error) {
	return "secondary", false, nil
}

type refundProc struct{}

func (refundProc) Name() string { return "refund" }
func (refundProc) Match(_ context.Context, input testInput) (bool, error) {
	return input.scene == "REFUND", nil
}
func (refundProc) Handle(_ context.Context, _ testInput) (string, bool, error) {
	return "refund", true, nil
}

type errMatchProc struct{}

func (errMatchProc) Name() string { return "err_match" }
func (errMatchProc) Match(_ context.Context, _ testInput) (bool, error) {
	return false, fmt.Errorf("match failed")
}
func (errMatchProc) Handle(_ context.Context, _ testInput) (string, bool, error) {
	return "", true, nil
}

type errHandleProc struct{}

func (errHandleProc) Name() string { return "err_handle" }
func (errHandleProc) Match(_ context.Context, _ testInput) (bool, error) {
	return true, nil
}
func (errHandleProc) Handle(_ context.Context, _ testInput) (string, bool, error) {
	return "", true, fmt.Errorf("handle failed")
}

func buildTemplate() Template[testProc, testInput, string] {
	return NewTemplate(
		func(ctx context.Context, impl testProc, input testInput) (bool, error) {
			return impl.Match(ctx, input)
		},
		func(ctx context.Context, impl testProc, input testInput) (string, bool, error) {
			return impl.Handle(ctx, input)
		},
	)
}

func TestTemplateSerialModeStopsByContinueFlag(t *testing.T) {
	template := buildTemplate()

	results, err := template(context.Background(), []testProc{
		primaryProc{},
		secondaryProc{},
		refundProc{},
	}, testInput{scene: "ORDER"}, Serial)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results before chain stop, got %d", len(results))
	}
	if results[0] != "primary" || results[1] != "secondary" {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestTemplateParallelMode(t *testing.T) {
	var mu sync.Mutex
	called := map[string]int{}

	template := NewTemplate(
		func(ctx context.Context, impl testProc, input testInput) (bool, error) {
			return impl.Match(ctx, input)
		},
		func(ctx context.Context, impl testProc, input testInput) (string, bool, error) {
			mu.Lock()
			called[impl.Name()]++
			mu.Unlock()
			return impl.Name(), false, nil
		},
	)

	results, err := template(context.Background(), []testProc{
		primaryProc{},
		secondaryProc{},
	}, testInput{scene: "ORDER"}, Parallel)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 parallel results, got %d", len(results))
	}
	if called["primary"] != 1 || called["secondary"] != 1 {
		t.Fatalf("unexpected call count: %#v", called)
	}
}

func TestTemplateDefaultModeIsSerial(t *testing.T) {
	template := buildTemplate()

	results, err := template(context.Background(), []testProc{
		primaryProc{},
		secondaryProc{},
	}, testInput{scene: "ORDER"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected serial behavior, got %d", len(results))
	}
}

func TestTemplateInvalidMode(t *testing.T) {
	template := buildTemplate()
	if _, err := template(context.Background(), nil, testInput{}, Mode("UNKNOWN")); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestTemplateNilProcess(t *testing.T) {
	template := NewTemplate[testProc, testInput, string](nil, nil)
	if _, err := template(context.Background(), nil, testInput{}, Serial); err == nil {
		t.Fatal("expected nil process error")
	}
}

func TestTemplateMatchError(t *testing.T) {
	template := buildTemplate()
	if _, err := template(context.Background(), []testProc{errMatchProc{}}, testInput{}, Serial); err == nil {
		t.Fatal("expected match error")
	}
}

func TestTemplateProcessError(t *testing.T) {
	template := buildTemplate()
	if _, err := template(context.Background(), []testProc{errHandleProc{}}, testInput{}, Serial); err == nil {
		t.Fatal("expected process error")
	}
}

func TestTemplateParallelJoinError(t *testing.T) {
	template := buildTemplate()
	_, err := template(context.Background(), []testProc{primaryProc{}, errHandleProc{}}, testInput{scene: "ORDER"}, Parallel)
	if err == nil {
		t.Fatal("expected parallel joined error")
	}
	if !strings.Contains(err.Error(), "handle failed") {
		t.Fatalf("expected joined handle failed error, got %v", err)
	}
}

func TestMergeImplementationsAppend(t *testing.T) {
	merged, err := MergeImplementations([]string{"a"}, []string{"b", "c"}, Append)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 3 || merged[0] != "a" || merged[1] != "b" || merged[2] != "c" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestMergeImplementationsAppendBefore(t *testing.T) {
	merged, err := MergeImplementationsWithAppendType([]string{"a"}, []string{"b", "c"}, Append, AppendBefore)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 3 || merged[0] != "b" || merged[1] != "c" || merged[2] != "a" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestMergeImplementationsAppendAfter(t *testing.T) {
	merged, err := MergeImplementationsWithAppendType([]string{"a"}, []string{"b", "c"}, Append, AppendAfter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 3 || merged[0] != "a" || merged[1] != "b" || merged[2] != "c" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestMergeImplementationsAppendParallel(t *testing.T) {
	merged, err := MergeImplementationsWithAppendType([]string{"a"}, []string{"b", "c"}, Append, AppendParallel)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 3 || merged[0] != "a" || merged[1] != "b" || merged[2] != "c" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestMergeImplementationsSkip(t *testing.T) {
	merged, err := MergeImplementations([]string{"a"}, []string{"b"}, Skip)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 1 || merged[0] != "a" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}

	merged, err = MergeImplementations(nil, []string{"b"}, Skip)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 1 || merged[0] != "b" {
		t.Fatalf("unexpected merged result without existing: %#v", merged)
	}
}

func TestMergeImplementationsOverwrite(t *testing.T) {
	merged, err := MergeImplementations([]string{"a"}, []string{"b", "c"}, Overwrite)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(merged) != 2 || merged[0] != "b" || merged[1] != "c" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}

func TestMergeImplementationsInvalidAction(t *testing.T) {
	if _, err := MergeImplementations([]string{"a"}, []string{"b"}, DefinitionAction("UNKNOWN")); err == nil {
		t.Fatal("expected invalid action error")
	}
}

func TestMergeImplementationsInvalidAppendType(t *testing.T) {
	if _, err := MergeImplementationsWithAppendType([]string{"a"}, []string{"b"}, Append, AppendType("UNKNOWN")); err == nil {
		t.Fatal("expected invalid append type error")
	}
}
