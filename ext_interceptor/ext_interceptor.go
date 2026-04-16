package ext_interceptor

import (
	"context"
	"fmt"
)

// Handler is the final business handler wrapped by interceptors.
type Handler[Input any, Output any] func(ctx context.Context, input Input) (Output, error)

// MatchFunc defines whether an interceptor should join current invocation.
type MatchFunc[Impl any, Input any] func(ctx context.Context, impl Impl, input Input) (bool, error)

// InterceptFunc defines how one interceptor wraps the next handler.
type InterceptFunc[Impl any, Input any, Output any] func(ctx context.Context, impl Impl, input Input, next Handler[Input, Output]) (Output, error)

// Template defines interceptor execution entry.
type Template[Impl any, Input any, Output any] func(ctx context.Context, interceptors []Impl, input Input, final Handler[Input, Output]) (Output, error)

// NewTemplate builds an interceptor template function.
func NewTemplate[Impl any, Input any, Output any](match MatchFunc[Impl, Input], intercept InterceptFunc[Impl, Input, Output]) Template[Impl, Input, Output] {
	return func(ctx context.Context, interceptors []Impl, input Input, final Handler[Input, Output]) (Output, error) {
		if intercept == nil {
			var zero Output
			return zero, fmt.Errorf("interceptor func is required")
		}
		if final == nil {
			var zero Output
			return zero, fmt.Errorf("final handler is required")
		}

		chain := final
		for i := len(interceptors) - 1; i >= 0; i-- {
			current := interceptors[i]
			if match != nil {
				matched, err := match(ctx, current, input)
				if err != nil {
					var zero Output
					return zero, err
				}
				if !matched {
					continue
				}
			}

			next := chain
			chain = func(ctx context.Context, input Input) (Output, error) {
				return intercept(ctx, current, input, next)
			}
		}

		return chain(ctx, input)
	}
}
