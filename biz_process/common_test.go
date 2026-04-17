package biz_process

import (
	"context"
	"testing"
)

func TestNodeImplementations(t *testing.T) {
	var task Node = Task{Name: "task"}
	var graphNode Node = GraphNode{Name: "graph"}
	var action Node = Action(func(ctx context.Context, from State, to State, event Event, payload any) error { return nil })

	if task.NodeName() != "task" {
		t.Fatalf("unexpected task node name: %s", task.NodeName())
	}
	if graphNode.NodeName() != "graph" {
		t.Fatalf("unexpected graph node name: %s", graphNode.NodeName())
	}
	if action.NodeName() != "action" {
		t.Fatalf("unexpected action node name: %s", action.NodeName())
	}
}
