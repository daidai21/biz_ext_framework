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

// DefinitionAction defines how incoming implementations change one ext process definition.
type DefinitionAction string

const (
	// Append appends incoming implementations after existing ones.
	Append DefinitionAction = "APPEND"
	// Skip keeps existing implementations and ignores incoming ones when the definition already exists.
	Skip DefinitionAction = "SKIP"
	// Overwrite replaces existing implementations with incoming ones.
	Overwrite DefinitionAction = "OVERWRITE"
)

// AppendType defines where incoming implementations are placed when action is Append.
type AppendType string

const (
	// AppendBefore prepends incoming implementations before existing ones.
	AppendBefore AppendType = "BEFORE"
	// AppendAfter appends incoming implementations after existing ones.
	AppendAfter AppendType = "AFTER"
	// AppendParallel appends incoming implementations after existing ones and is intended
	// to be combined with Execute(..., Parallel) when the flow should run concurrently.
	AppendParallel AppendType = "PARALLEL"
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

// Validate checks whether the definition action is supported.
func (a DefinitionAction) Validate() error {
	switch a {
	case Append, Skip, Overwrite:
		return nil
	default:
		return fmt.Errorf("unsupported ext process definition action: %q", string(a))
	}
}

func normalizeDefinitionAction(action DefinitionAction) DefinitionAction {
	if action == "" {
		return Append
	}
	return action
}

// Validate checks whether the append type is supported.
func (t AppendType) Validate() error {
	switch t {
	case AppendBefore, AppendAfter, AppendParallel:
		return nil
	default:
		return fmt.Errorf("unsupported ext process append type: %q", string(t))
	}
}

func normalizeAppendType(appendType AppendType) AppendType {
	if appendType == "" {
		return AppendAfter
	}
	return appendType
}

// MergeImplementations applies the action to existing and incoming implementations.
func MergeImplementations[Impl any](existing []Impl, incoming []Impl, action DefinitionAction) ([]Impl, error) {
	return MergeImplementationsWithAppendType(existing, incoming, action, AppendAfter)
}

// MergeImplementationsWithAppendType applies the action and append type to existing and incoming implementations.
func MergeImplementationsWithAppendType[Impl any](existing []Impl, incoming []Impl, action DefinitionAction, appendType AppendType) ([]Impl, error) {
	action = normalizeDefinitionAction(action)
	if err := action.Validate(); err != nil {
		return nil, err
	}
	appendType = normalizeAppendType(appendType)
	if err := appendType.Validate(); err != nil {
		return nil, err
	}

	switch action {
	case Append:
		switch appendType {
		case AppendBefore:
			merged := append([]Impl(nil), incoming...)
			merged = append(merged, existing...)
			return merged, nil
		case AppendAfter, AppendParallel:
			merged := append([]Impl(nil), existing...)
			merged = append(merged, incoming...)
			return merged, nil
		default:
			return nil, fmt.Errorf("unsupported ext process append type: %q", appendType)
		}
	case Skip:
		if len(existing) > 0 {
			return append([]Impl(nil), existing...), nil
		}
		return append([]Impl(nil), incoming...), nil
	case Overwrite:
		return append([]Impl(nil), incoming...), nil
	default:
		return nil, fmt.Errorf("unsupported ext process definition action: %q", action)
	}
}
