package biz_process

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestRunDAGSerialChain(t *testing.T) {
	var mu sync.Mutex
	order := make([]string, 0, 3)
	appendOrder := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, name)
	}

	dag := []DAGNode{
		{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return nil }},
		{Name: "B", DependsOn: []string{"A"}, Task: func(ctx context.Context) error { appendOrder("B"); return nil }},
		{Name: "C", DependsOn: []string{"B"}, Task: func(ctx context.Context) error { appendOrder("C"); return nil }},
	}

	if err := RunDAG(context.Background(), dag); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if strings.Join(order, ",") != "A,B,C" {
		t.Fatalf("expected A,B,C, got %v", order)
	}
}

func TestRunDAGParallelFanoutJoin(t *testing.T) {
	var mu sync.Mutex
	counter := map[string]int{}
	inc := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		counter[name]++
	}

	dag := []DAGNode{
		{Name: "prepare", Task: func(ctx context.Context) error { inc("prepare"); return nil }},
		{Name: "p1", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { inc("p1"); return nil }},
		{Name: "p2", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { inc("p2"); return nil }},
		{Name: "merge", DependsOn: []string{"p1", "p2"}, Task: func(ctx context.Context) error { inc("merge"); return nil }},
	}

	if err := RunDAG(context.Background(), dag); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	for _, key := range []string{"prepare", "p1", "p2", "merge"} {
		if counter[key] != 1 {
			t.Fatalf("expected %s execute once, got %d", key, counter[key])
		}
	}
}

func TestRunDAGEmpty(t *testing.T) {
	err := RunDAG(context.Background(), nil)
	if !errors.Is(err, ErrEmptyDAG) {
		t.Fatalf("expected ErrEmptyDAG, got %v", err)
	}
}

func TestRunDAGInvalidDependency(t *testing.T) {
	dag := []DAGNode{
		{Name: "A", DependsOn: []string{"X"}, Task: func(ctx context.Context) error { return nil }},
	}
	err := RunDAG(context.Background(), dag)
	if !errors.Is(err, ErrInvalidDAG) {
		t.Fatalf("expected ErrInvalidDAG, got %v", err)
	}
}

func TestRunDAGDuplicateName(t *testing.T) {
	dag := []DAGNode{
		{Name: "A", Task: func(ctx context.Context) error { return nil }},
		{Name: "A", Task: func(ctx context.Context) error { return nil }},
	}
	err := RunDAG(context.Background(), dag)
	if !errors.Is(err, ErrInvalidDAG) {
		t.Fatalf("expected ErrInvalidDAG, got %v", err)
	}
}

func TestRunDAGCycle(t *testing.T) {
	dag := []DAGNode{
		{Name: "A", DependsOn: []string{"B"}, Task: func(ctx context.Context) error { return nil }},
		{Name: "B", DependsOn: []string{"A"}, Task: func(ctx context.Context) error { return nil }},
	}
	err := RunDAG(context.Background(), dag)
	if !errors.Is(err, ErrDAGCycle) {
		t.Fatalf("expected ErrDAGCycle, got %v", err)
	}
}

func TestRunDAGTaskErrorStopsDownstream(t *testing.T) {
	var mu sync.Mutex
	order := make([]string, 0, 2)
	appendOrder := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, name)
	}

	dag := []DAGNode{
		{Name: "A", Task: func(ctx context.Context) error { appendOrder("A"); return errors.New("A failed") }},
		{Name: "B", DependsOn: []string{"A"}, Task: func(ctx context.Context) error { appendOrder("B"); return nil }},
	}

	err := RunDAG(context.Background(), dag)
	if err == nil {
		t.Fatal("expected dag fail")
	}
	if !strings.Contains(err.Error(), "A failed") {
		t.Fatalf("expected A failed message, got %v", err)
	}
	if strings.Join(order, ",") != "A" {
		t.Fatalf("expected B not executed, got %v", order)
	}
}
