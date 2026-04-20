package biz_process

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRerunnerExecuteSuccessFirstAttempt(t *testing.T) {
	var called atomic.Int32
	rerunner := Rerunner[string, string]{Attempts: 3}

	value, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		called.Add(1)
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if value != "ok" {
		t.Fatalf("unexpected value: %q", value)
	}
	if called.Load() != 1 {
		t.Fatalf("expected one call, got %d", called.Load())
	}
}

func TestRerunnerExecuteRerunsUntilSuccess(t *testing.T) {
	var called atomic.Int32
	rerunner := Rerunner[string, string]{Attempts: 3}

	value, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		if called.Add(1) < 3 {
			return "", errors.New("retry")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if value != "ok" {
		t.Fatalf("unexpected value: %q", value)
	}
	if called.Load() != 3 {
		t.Fatalf("expected three calls, got %d", called.Load())
	}
}

func TestRerunnerExecuteReturnsLastError(t *testing.T) {
	var called atomic.Int32
	expectedErr := errors.New("still failed")
	rerunner := Rerunner[string, string]{Attempts: 3}

	_, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		called.Add(1)
		return "", expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected last error, got %v", err)
	}
	if called.Load() != 3 {
		t.Fatalf("expected three calls, got %d", called.Load())
	}
}

func TestRerunnerExecuteHonorsShouldRerun(t *testing.T) {
	var called atomic.Int32
	expectedErr := errors.New("stop")
	rerunner := Rerunner[string, string]{
		Attempts: 5,
		ShouldRerun: func(context.Context, int, error) bool {
			return false
		},
	}

	_, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		called.Add(1)
		return "", expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected stop error, got %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("expected one call, got %d", called.Load())
	}
}

func TestRerunnerExecuteCallsOnRerun(t *testing.T) {
	var nextAttempt atomic.Int32
	rerunner := Rerunner[string, string]{
		Attempts: 2,
		OnRerun: func(_ context.Context, attempt int, err error) {
			if err == nil {
				t.Fatal("expected rerun hook error")
			}
			nextAttempt.Store(int32(attempt))
		},
	}

	_, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		return "", errors.New("retry")
	})
	if err == nil {
		t.Fatal("expected execute error")
	}
	if nextAttempt.Load() != 2 {
		t.Fatalf("expected next attempt 2, got %d", nextAttempt.Load())
	}
}

func TestRerunnerExecuteCancelsDuringInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rerunner := Rerunner[string, string]{
		Attempts: 2,
		Interval: 50 * time.Millisecond,
		OnRerun: func(context.Context, int, error) {
			cancel()
		},
	}

	_, err := rerunner.Execute(ctx, "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		return "", errors.New("retry")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestRerunnerExecuteDefaultsAttemptsToOne(t *testing.T) {
	var called atomic.Int32
	rerunner := Rerunner[string, string]{}

	_, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		called.Add(1)
		return "", errors.New("failed")
	})
	if err == nil {
		t.Fatal("expected execute error")
	}
	if called.Load() != 1 {
		t.Fatalf("expected one call, got %d", called.Load())
	}
}

func TestRerunnerExecuteValidatesConfig(t *testing.T) {
	rerunner := Rerunner[string, string]{Attempts: -1}

	_, err := rerunner.Execute(context.Background(), "req-1", func(_ context.Context, req string) (string, error) {
		if req != "req-1" {
			t.Fatalf("unexpected req: %q", req)
		}
		return "ok", nil
	})
	if !errors.Is(err, ErrInvalidRerunAttempts) {
		t.Fatalf("expected ErrInvalidRerunAttempts, got %v", err)
	}

	_, err = rerunner.Execute(context.Background(), "req-1", nil)
	if !errors.Is(err, ErrNilRerunFunc) {
		t.Fatalf("expected ErrNilRerunFunc, got %v", err)
	}
}
