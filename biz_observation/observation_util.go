package biz_observation

import (
	"context"
	"slices"
)

type DependencyLevel string

const (
	DependencyLevelStrong DependencyLevel = "STRONG"
	DependencyLevelWeak   DependencyLevel = "WEAK"
)

type dependencyContextKey struct{}

type DependencyMap map[string]DependencyLevel

func WithDependency(ctx context.Context, name string, level DependencyLevel) context.Context {
	if ctx == nil || name == "" {
		return ctx
	}

	dependencies := DependenciesFromContext(ctx)
	if dependencies == nil {
		dependencies = make(DependencyMap)
	}

	if existing, ok := dependencies[name]; ok {
		if existing == DependencyLevelStrong || level == DependencyLevelWeak {
			return context.WithValue(ctx, dependencyContextKey{}, dependencies)
		}
	}

	dependencies[name] = level
	return context.WithValue(ctx, dependencyContextKey{}, dependencies)
}

func WithStrongDependency(ctx context.Context, name string) context.Context {
	return WithDependency(ctx, name, DependencyLevelStrong)
}

func WithWeakDependency(ctx context.Context, name string) context.Context {
	return WithDependency(ctx, name, DependencyLevelWeak)
}

func DependenciesFromContext(ctx context.Context) DependencyMap {
	if ctx == nil {
		return nil
	}
	dependencies, _ := ctx.Value(dependencyContextKey{}).(DependencyMap)
	if dependencies == nil {
		return nil
	}

	copied := make(DependencyMap, len(dependencies))
	for name, level := range dependencies {
		copied[name] = level
	}
	return copied
}

func DependencyLevelFromContext(ctx context.Context, name string) (DependencyLevel, bool) {
	dependencies := DependenciesFromContext(ctx)
	level, ok := dependencies[name]
	return level, ok
}

func StrongDependenciesFromContext(ctx context.Context) []string {
	return dependenciesByLevel(ctx, DependencyLevelStrong)
}

func WeakDependenciesFromContext(ctx context.Context) []string {
	return dependenciesByLevel(ctx, DependencyLevelWeak)
}

func dependenciesByLevel(ctx context.Context, level DependencyLevel) []string {
	dependencies := DependenciesFromContext(ctx)
	names := make([]string, 0, len(dependencies))
	for name, current := range dependencies {
		if current == level {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}
