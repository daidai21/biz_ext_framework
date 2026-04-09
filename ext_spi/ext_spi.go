package ext_spi

import (
	"context"
	"fmt"
)

// Mode defines how a SPI template resolves matched implementations.
type Mode string

const (
	// First invokes only the first implementation.
	First Mode = "FIRST"
	// All invokes every implementation.
	All Mode = "ALL"
	// FirstMatched invokes only the first matched implementation.
	FirstMatched Mode = "FIRST_MATCHED"
	// AllMatched invokes every matched implementation.
	AllMatched Mode = "ALL_MATCHED"
)

// MatchFunc defines how a SPI implementation is matched before invocation.
type MatchFunc[Impl any, Input any] func(ctx context.Context, impl Impl, input Input) (bool, error)

// InvokeFunc defines how a SPI implementation is invoked.
type InvokeFunc[Impl any, Input any, Output any] func(ctx context.Context, impl Impl, input Input) (Output, error)

// Template defines the SPI call entry.
type Template[Impl any, Input any, Output any] func(ctx context.Context, extSpiImpls []Impl, input Input, mode Mode) ([]Output, error)

// NewTemplate builds a SPI template function.
func NewTemplate[Impl any, Input any, Output any](match MatchFunc[Impl, Input], invoke InvokeFunc[Impl, Input, Output]) Template[Impl, Input, Output] {
	return func(ctx context.Context, extSpiImpls []Impl, input Input, mode Mode) ([]Output, error) {
		if invoke == nil {
			return nil, fmt.Errorf("spi invoke func is required")
		}

		mode = normalizeMode(mode)
		if err := mode.Validate(); err != nil {
			return nil, err
		}
		if mode.requiresMatch() && match == nil {
			return nil, fmt.Errorf("spi match func is required for matched modes")
		}

		results := make([]Output, 0, len(extSpiImpls))
		for _, impl := range extSpiImpls {
			if mode.requiresMatch() {
				matched, err := match(ctx, impl, input)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}

			output, err := invoke(ctx, impl, input)
			if err != nil {
				return nil, err
			}
			results = append(results, output)
			if mode.isFirst() {
				break
			}
		}

		return results, nil
	}
}

// Validate checks whether the mode is supported.
func (m Mode) Validate() error {
	switch m {
	case First, All, FirstMatched, AllMatched:
		return nil
	default:
		return fmt.Errorf("unsupported spi mode: %q", string(m))
	}
}

func (m Mode) isFirst() bool {
	switch m {
	case First, FirstMatched:
		return true
	default:
		return false
	}
}

func (m Mode) requiresMatch() bool {
	switch m {
	case FirstMatched, AllMatched:
		return true
	default:
		return false
	}
}

func normalizeMode(mode Mode) Mode {
	if mode == "" {
		return First
	}
	return mode
}
