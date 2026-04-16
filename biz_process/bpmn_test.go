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

	process := Process{
		Layers: []ProcessLayer{
			{Name: "layer-a", Nodes: []ProcessNode{TaskProcessNode{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return nil }}}},
			{Name: "layer-b", Nodes: []ProcessNode{TaskProcessNode{Name: "B", Task: func(ctx context.Context) error { appendOrder("B"); return nil }}}},
			{Name: "layer-c", Nodes: []ProcessNode{TaskProcessNode{Name: "C", Task: func(ctx context.Context) error { appendOrder("C"); return nil }}}},
		},
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

	process := Process{
		Layers: []ProcessLayer{
			{Name: "prepare", Nodes: []ProcessNode{TaskProcessNode{Name: "prepare", Task: func(ctx context.Context) error { inc("prepare"); return nil }}}},
			{Name: "fanout", Nodes: []ProcessNode{
				TaskProcessNode{Name: "p1", Task: func(ctx context.Context) error { inc("p1"); return nil }},
				TaskProcessNode{Name: "p2", Task: func(ctx context.Context) error { inc("p2"); return nil }},
			}},
			{Name: "merge", Nodes: []ProcessNode{TaskProcessNode{Name: "merge", Task: func(ctx context.Context) error { inc("merge"); return nil }}}},
		},
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
	err := RunProcess(context.Background(), Process{})
	if !errors.Is(err, ErrEmptyProcess) {
		t.Fatalf("expected ErrEmptyProcess, got %v", err)
	}
}

func TestRunProcessInvalidProcess(t *testing.T) {
	process := Process{
		Layers: []ProcessLayer{
			{Name: "bad"},
		},
	}
	err := RunProcess(context.Background(), process)
	if !errors.Is(err, ErrInvalidProcess) {
		t.Fatalf("expected ErrInvalidProcess, got %v", err)
	}
}

func TestRunProcessInvalidNilNode(t *testing.T) {
	process := Process{
		Layers: []ProcessLayer{
			{Name: "bad", Nodes: []ProcessNode{nil}},
		},
	}

	err := RunProcess(context.Background(), process)
	if !errors.Is(err, ErrInvalidProcess) {
		t.Fatalf("expected ErrInvalidProcess, got %v", err)
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

	process := Process{
		Layers: []ProcessLayer{
			{Name: "layer-a", Nodes: []ProcessNode{TaskProcessNode{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return nil }}}},
			{Name: "layer-b", Nodes: []ProcessNode{TaskProcessNode{Name: "B", Task: func(ctx context.Context) error { appendOrder("B"); return errors.New("B failed") }}}},
			{Name: "layer-c", Nodes: []ProcessNode{TaskProcessNode{Name: "C", Task: func(ctx context.Context) error { appendOrder("C"); return nil }}}},
		},
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
	process := Process{
		Layers: []ProcessLayer{
			{Name: "fanout", Nodes: []ProcessNode{
				TaskProcessNode{Name: "ok", Task: func(ctx context.Context) error { return nil }},
				TaskProcessNode{Name: "bad", Task: func(ctx context.Context) error { return errors.New("branch failed") }},
			}},
		},
	}

	err := RunProcess(context.Background(), process)
	if err == nil {
		t.Fatal("expected parallel error")
	}
	if !strings.Contains(err.Error(), "parallel layer") {
		t.Fatalf("expected parallel layer message, got %v", err)
	}
	if !strings.Contains(err.Error(), "branch failed") {
		t.Fatalf("expected branch failed message, got %v", err)
	}
}
