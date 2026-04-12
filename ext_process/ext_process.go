package ext_process

import (
	"context"
	"errors"
	"fmt"
)

// Mode defines how ext process implementations are executed.
type Mode string

const (
	// Serial executes matched implementations in registration order.
	// If continueNext=false is returned by one implementation, the chain stops.
	Serial Mode = "SERIAL"
	// Parallel executes matched implementations concurrently.
	// continueNext is ignored in this mode.
	Parallel Mode = "PARALLEL"
)

// MatchFunc defines whether an implementation should join current process call.
type MatchFunc[Impl any, Input any] func(ctx context.Context, impl Impl, input Input) (bool, error)

// ProcessFunc defines how one implementation processes input.
// continueNext only affects Serial mode.
type ProcessFunc[Impl any, Input any, Output any] func(ctx context.Context, impl Impl, input Input) (output Output, continueNext bool, err error)

// Template defines ext process call entry.
type Template[Impl any, Input any, Output any] func(ctx context.Context, extProcessImpls []Impl, input Input, mode Mode) ([]Output, error)

// NewTemplate builds an ext process template function.
func NewTemplate[Impl any, Input any, Output any](match MatchFunc[Impl, Input], process ProcessFunc[Impl, Input, Output]) Template[Impl, Input, Output] {
	return func(ctx context.Context, extProcessImpls []Impl, input Input, mode Mode) ([]Output, error) {
		if process == nil {
			return nil, fmt.Errorf("ext process func is required")
		}

		mode = normalizeMode(mode)
		if err := mode.Validate(); err != nil {
			return nil, err
		}

		type selectedImpl struct {
			impl Impl
		}
		selected := make([]selectedImpl, 0, len(extProcessImpls))
		for _, impl := range extProcessImpls {
			if match != nil {
				matched, err := match(ctx, impl, input)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}
			selected = append(selected, selectedImpl{impl: impl})
		}

		switch mode {
		case Serial:
			results := make([]Output, 0, len(selected))
			for _, item := range selected {
				output, continueNext, err := process(ctx, item.impl, input)
				if err != nil {
					return nil, err
				}
				results = append(results, output)
				if !continueNext {
					break
				}
			}
			return results, nil
		case Parallel:
			type processResult struct {
				idx    int
				output Output
				err    error
			}
			ch := make(chan processResult, len(selected))
			for i, item := range selected {
				idx := i
				impl := item.impl
				go func() {
					output, _, err := process(ctx, impl, input)
					ch <- processResult{idx: idx, output: output, err: err}
				}()
			}

			ordered := make([]Output, len(selected))
			ok := make([]bool, len(selected))
			var errs []error
			for i := 0; i < len(selected); i++ {
				result := <-ch
				if result.err != nil {
					errs = append(errs, result.err)
					continue
				}
				ordered[result.idx] = result.output
				ok[result.idx] = true
			}
			if len(errs) > 0 {
				return nil, errors.Join(errs...)
			}
			results := make([]Output, 0, len(selected))
			for i := 0; i < len(ordered); i++ {
				if ok[i] {
					results = append(results, ordered[i])
				}
			}
			return results, nil
		default:
			return nil, fmt.Errorf("unsupported ext process mode: %q", mode)
		}
	}
}

// Validate checks whether the mode is supported.
func (m Mode) Validate() error {
	switch m {
	case Serial, Parallel:
		return nil
	default:
		return fmt.Errorf("unsupported ext process mode: %q", string(m))
	}
}

func normalizeMode(mode Mode) Mode {
	if mode == "" {
		return Serial
	}
	return mode
}
