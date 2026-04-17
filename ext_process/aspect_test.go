package ext_process

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAspectNoop(t *testing.T) {
	if err := Aspect(context.Background(), testInput{scene: "ORDER"}); err != nil {
		t.Fatalf("expected noop aspect success, got %v", err)
	}
	if err := Aspect(nil, testInput{scene: "ORDER"}); err != nil {
		t.Fatalf("expected nil ctx aspect success, got %v", err)
	}
}

func TestWithAspect(t *testing.T) {
	called := 0
	ctx := WithAspect(context.Background(), func(ctx context.Context, input any) error {
		called++
		typed, ok := input.(testInput)
		if !ok || typed.scene != "ORDER" {
			t.Fatalf("unexpected aspect input: %#v", input)
		}
		return nil
	})

	if err := Aspect(ctx, testInput{scene: "ORDER"}); err != nil {
		t.Fatalf("expected aspect success, got %v", err)
	}
	if called != 1 {
		t.Fatalf("expected aspect called once, got %d", called)
	}

	if same := WithAspect(context.Background(), nil); same == nil {
		t.Fatal("expected nil runner registration to keep context")
	}
}

func TestBindAspect(t *testing.T) {
	template := buildTemplate()
	ctx := BindAspect(context.Background(), template, []testProc{
		primaryProc{},
		secondaryProc{},
	}, Serial)

	if err := Aspect(ctx, testInput{scene: "ORDER"}); err != nil {
		t.Fatalf("expected bound aspect success, got %v", err)
	}
}

func TestBindAspectNilTemplate(t *testing.T) {
	ctx := BindAspect[testProc, testInput, string](context.Background(), nil, nil, Serial)
	if err := Aspect(ctx, testInput{scene: "ORDER"}); !errors.Is(err, ErrNilAspectTemplate) {
		t.Fatalf("expected ErrNilAspectTemplate, got %v", err)
	}
}

func TestBindAspectTypeMismatch(t *testing.T) {
	ctx := BindAspect(context.Background(), buildTemplate(), []testProc{primaryProc{}}, Serial)
	if err := Aspect(ctx, "bad-input"); !errors.Is(err, ErrAspectInputTypeMismatch) {
		t.Fatalf("expected ErrAspectInputTypeMismatch, got %v", err)
	}
}

func TestAspectJoinErrors(t *testing.T) {
	ctx := WithAspect(context.Background(), func(context.Context, any) error {
		return errors.New("first")
	})
	ctx = WithAspect(ctx, func(context.Context, any) error {
		return errors.New("second")
	})

	err := Aspect(ctx, testInput{scene: "ORDER"})
	if err == nil {
		t.Fatal("expected joined aspect error")
	}
	if got := err.Error(); !strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Fatalf("expected joined error contains both messages, got %v", err)
	}
	if got := err.Error(); got == "first" || got == "second" {
		t.Fatalf("expected joined error, got %v", err)
	}
}
