package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_process"
)

func TestProcessContainerRun(t *testing.T) {
	var order []string
	process := biz_process.Process{
		Layers: []biz_process.ProcessLayer{
			{
				Name: "prepare",
				Nodes: []biz_process.ProcessNode{
					biz_process.TaskProcessNode{Name: "prepare", Task: func(ctx context.Context) error {
						order = append(order, "prepare")
						return nil
					}},
				},
			},
			{
				Name: "finalize",
				Nodes: []biz_process.ProcessNode{
					biz_process.TaskProcessNode{Name: "finalize", Task: func(ctx context.Context) error {
						order = append(order, "finalize")
						return nil
					}},
				},
			},
		},
	}

	container := NewProcessContainer()
	if err := container.Register("order_flow", process); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := container.Run(context.Background(), "order_flow"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(order) != 2 || order[0] != "prepare" || order[1] != "finalize" {
		t.Fatalf("unexpected process order: %v", order)
	}
}

func TestProcessContainerRunNotFound(t *testing.T) {
	container := NewProcessContainer()
	err := container.Run(context.Background(), "missing")
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
}
