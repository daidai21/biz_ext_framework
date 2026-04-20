package biz_process

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidRerunAttempts = errors.New("invalid rerun attempts")
	ErrNilRerunFunc         = errors.New("nil rerun func")
)

// ShouldRerunFunc decides whether Execute should continue with the next attempt.
type ShouldRerunFunc func(ctx context.Context, attempt int, err error) bool

// OnRerunFunc runs after one failed attempt and before the next attempt starts.
type OnRerunFunc func(ctx context.Context, nextAttempt int, err error)

// Rerunner executes one function with bounded rerun capability.
type Rerunner[Req any, Resp any] struct {
	// Attempts is the total number of executions, including the first one.
	// Zero means 1. Negative values are invalid.
	Attempts int
	// Interval waits between failed attempts. Zero means no wait.
	Interval time.Duration
	// ShouldRerun customizes whether a failed attempt should rerun.
	// Nil means rerun until Attempts is exhausted.
	ShouldRerun ShouldRerunFunc
	// OnRerun is an optional hook invoked before the next attempt starts.
	OnRerun OnRerunFunc
}

// Execute runs fn and reruns it based on the configured policy.
func (r Rerunner[Req, Resp]) Execute(ctx context.Context, req Req, fn func(context.Context, Req) (Resp, error)) (Resp, error) {
	var zero Resp
	if fn == nil {
		return zero, ErrNilRerunFunc
	}

	attempts := r.Attempts
	switch {
	case attempts == 0:
		attempts = 1
	case attempts < 0:
		return zero, ErrInvalidRerunAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := fn(ctx, req)
		if err == nil {
			return value, nil
		}
		lastErr = err

		if attempt == attempts {
			break
		}
		if r.ShouldRerun != nil && !r.ShouldRerun(ctx, attempt, err) {
			break
		}
		if r.OnRerun != nil {
			r.OnRerun(ctx, attempt+1, err)
		}
		if waitErr := waitRerunInterval(ctx, r.Interval); waitErr != nil {
			return zero, waitErr
		}
	}

	return zero, lastErr
}

func waitRerunInterval(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
