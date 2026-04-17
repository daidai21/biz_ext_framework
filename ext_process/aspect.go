package ext_process

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNilAspectTemplate       = errors.New("nil ext process aspect template")
	ErrAspectInputTypeMismatch = errors.New("ext process aspect input type mismatch")
)

type aspectRunner func(ctx context.Context, input any) error

type aspectContextKey struct{}

// WithAspect registers one aspect runner into context.
func WithAspect(ctx context.Context, runner func(ctx context.Context, input any) error) context.Context {
	if ctx == nil || runner == nil {
		return ctx
	}

	runners := aspectRunnersFromContext(ctx)
	runners = append(runners, runner)
	return context.WithValue(ctx, aspectContextKey{}, runners)
}

// BindAspect binds one typed ext_process template into context so business code
// can trigger it later via defer Aspect(ctx, input).
func BindAspect[Impl any, Input any, Output any](ctx context.Context, template Template[Impl, Input, Output], extProcessImpls []Impl, mode Mode) context.Context {
	if template == nil {
		return WithAspect(ctx, func(context.Context, any) error {
			return ErrNilAspectTemplate
		})
	}

	impls := append([]Impl(nil), extProcessImpls...)
	return WithAspect(ctx, func(ctx context.Context, input any) error {
		typed, ok := input.(Input)
		if !ok {
			return fmt.Errorf("%w: got %T", ErrAspectInputTypeMismatch, input)
		}
		_, err := template(ctx, impls, typed, mode)
		return err
	})
}

// Aspect executes all aspect runners bound in context.
// It is designed for business code to use with defer:
// defer ext_process.Aspect(ctx, input)
func Aspect(ctx context.Context, input any) error {
	runners := aspectRunnersFromContext(ctx)
	if len(runners) == 0 {
		return nil
	}

	var errs []error
	for _, runner := range runners {
		if err := runner(ctx, input); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func aspectRunnersFromContext(ctx context.Context) []aspectRunner {
	if ctx == nil {
		return nil
	}

	runners, _ := ctx.Value(aspectContextKey{}).([]aspectRunner)
	return append([]aspectRunner(nil), runners...)
}
