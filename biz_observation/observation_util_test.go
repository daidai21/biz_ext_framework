package biz_observation

import "context"
import "testing"

func TestWithDependency(t *testing.T) {
	ctx := context.Background()
	ctx = WithWeakDependency(ctx, "metrics")
	ctx = WithStrongDependency(ctx, "trace")
	ctx = WithStrongDependency(ctx, "metrics")
	ctx = WithWeakDependency(ctx, "trace")

	dependencies := DependenciesFromContext(ctx)
	if len(dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(dependencies))
	}

	level, ok := DependencyLevelFromContext(ctx, "metrics")
	if !ok || level != DependencyLevelStrong {
		t.Fatalf("expected metrics strong, got %v %v", level, ok)
	}

	level, ok = DependencyLevelFromContext(ctx, "trace")
	if !ok || level != DependencyLevelStrong {
		t.Fatalf("expected trace strong, got %v %v", level, ok)
	}
}

func TestDependencyLists(t *testing.T) {
	ctx := context.Background()
	ctx = WithStrongDependency(ctx, "logger")
	ctx = WithWeakDependency(ctx, "metrics")

	strong := StrongDependenciesFromContext(ctx)
	weak := WeakDependenciesFromContext(ctx)

	if len(strong) != 1 || strong[0] != "logger" {
		t.Fatalf("unexpected strong deps: %v", strong)
	}
	if len(weak) != 1 || weak[0] != "metrics" {
		t.Fatalf("unexpected weak deps: %v", weak)
	}
}
