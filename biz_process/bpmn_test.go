package biz_process

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestRunProcessSerial(t *testing.T) {
	var mu sync.Mutex
	order := make([]string, 0, 3)
	appendOrder := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, name)
	}

	process := []Step{
		{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return nil }},
		{Name: "B", Task: func(ctx context.Context) error { appendOrder("B"); return nil }},
		{Name: "C", Task: func(ctx context.Context) error { appendOrder("C"); return nil }},
	}

	if err := RunProcess(context.Background(), process); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	got := strings.Join(order, ",")
	if got != "A,B,C" {
		t.Fatalf("expected serial order A,B,C, got %s", got)
	}
}

func TestRunProcessSerialWithParallelGroup(t *testing.T) {
	var mu sync.Mutex
	counter := map[string]int{}
	inc := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		counter[name]++
	}

	process := []Step{
		{Name: "prepare", Task: func(ctx context.Context) error { inc("prepare"); return nil }},
		{Name: "fanout", Parallel: []Step{
			{Name: "p1", Task: func(ctx context.Context) error { inc("p1"); return nil }},
			{Name: "p2", Task: func(ctx context.Context) error { inc("p2"); return nil }},
		}},
		{Name: "merge", Task: func(ctx context.Context) error { inc("merge"); return nil }},
	}

	if err := RunProcess(context.Background(), process); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	for _, key := range []string{"prepare", "p1", "p2", "merge"} {
		if counter[key] != 1 {
			t.Fatalf("expected %s execute once, got %d", key, counter[key])
		}
	}
}

func TestRunProcessEmpty(t *testing.T) {
	err := RunProcess(context.Background(), nil)
	if !errors.Is(err, ErrEmptyProcess) {
		t.Fatalf("expected ErrEmptyProcess, got %v", err)
	}
}

func TestRunProcessInvalidStep(t *testing.T) {
	process := []Step{{Name: "bad"}}
	err := RunProcess(context.Background(), process)
	if !errors.Is(err, ErrInvalidStep) {
		t.Fatalf("expected ErrInvalidStep, got %v", err)
	}
}

func TestRunProcessStepErrorStopsPipeline(t *testing.T) {
	var mu sync.Mutex
	order := make([]string, 0, 2)
	appendOrder := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, name)
	}

	process := []Step{
		{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return nil }},
		{Name: "B", Task: func(ctx context.Context) error { appendOrder("B"); return errors.New("B failed") }},
		{Name: "C", Task: func(ctx context.Context) error { appendOrder("C"); return nil }},
	}

	err := RunProcess(context.Background(), process)
	if err == nil {
		t.Fatal("expected process fail")
	}
	if !strings.Contains(err.Error(), "B") {
		t.Fatalf("expected error contains step name B, got %v", err)
	}
	if strings.Join(order, ",") != "A,B" {
		t.Fatalf("expected C not executed, got %v", order)
	}
}

func TestRunProcessParallelError(t *testing.T) {
	process := []Step{
		{Name: "fanout", Parallel: []Step{
			{Name: "ok", Task: func(ctx context.Context) error { return nil }},
			{Name: "bad", Task: func(ctx context.Context) error { return errors.New("branch failed") }},
		}},
	}

	err := RunProcess(context.Background(), process)
	if err == nil {
		t.Fatal("expected parallel error")
	}
	if !strings.Contains(err.Error(), "parallel group") {
		t.Fatalf("expected parallel group message, got %v", err)
	}
	if !strings.Contains(err.Error(), "branch failed") {
		t.Fatalf("expected branch failed message, got %v", err)
	}
}
