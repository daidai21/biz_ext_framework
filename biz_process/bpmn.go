package biz_process

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrEmptyProcess   = errors.New("empty process")
	ErrInvalidProcess = errors.New("invalid process")
)

// TaskFunc is one executable process step.
type TaskFunc func(ctx context.Context) error

// Process is a lightweight BPMN-like process.
// Layers run in serial order; nodes in the same layer run in parallel.
type Process struct {
	Name   string
	Layers []ProcessLayer
}

// ProcessLayer is one serial stage in a process.
type ProcessLayer struct {
	Name  string
	Nodes []ProcessNode
}

// ProcessNode is one executable BPMN node inside a process layer.
type ProcessNode interface {
	Node
	Run(ctx context.Context) error
}

// Task is the default BPMN node implementation backed by TaskFunc.
type Task struct {
	Name string
	Task TaskFunc
}

// TaskProcessNode is kept as a compatibility alias.
type TaskProcessNode = Task

func (n Task) NodeName() string {
	return n.Name
}

func (n Task) Run(ctx context.Context) error {
	if n.Task == nil {
		return fmt.Errorf("%w: node %q task is required", ErrInvalidProcess, n.Name)
	}
	return n.Task(ctx)
}

// RunProcess executes a process with serial layers and parallel nodes.
func RunProcess(ctx context.Context, process Process) error {
	if len(process.Layers) == 0 {
		return ErrEmptyProcess
	}

	for i, layer := range process.Layers {
		if err := runLayer(ctx, i, layer); err != nil {
			return fmt.Errorf("process layer[%d] %q failed: %w", i, layer.Name, err)
		}
	}
	return nil
}

func runLayer(ctx context.Context, layerIndex int, layer ProcessLayer) error {
	if len(layer.Nodes) == 0 {
		return fmt.Errorf("%w: layer[%d] %q must define at least one node", ErrInvalidProcess, layerIndex, layer.Name)
	}

	if len(layer.Nodes) == 1 {
		node := layer.Nodes[0]
		if err := runNode(ctx, layerIndex, 0, node); err != nil {
			return err
		}
		return nil
	}

	return runParallel(ctx, layerIndex, layer.Name, layer.Nodes)
}

func runNode(ctx context.Context, layerIndex, nodeIndex int, node ProcessNode) error {
	if node == nil {
		return fmt.Errorf("%w: layer[%d] node[%d] is nil", ErrInvalidProcess, layerIndex, nodeIndex)
	}

	nodeName := node.NodeName()
	if err := node.Run(ctx); err != nil {
		return fmt.Errorf("node %q failed: %w", nodeName, err)
	}
	return nil
}

func runParallel(ctx context.Context, layerIndex int, layerName string, nodes []ProcessNode) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(nodes))

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, current ProcessNode) {
			defer wg.Done()
			if err := runNode(ctx, layerIndex, idx, current); err != nil {
				nodeName := ""
				if current != nil {
					nodeName = current.NodeName()
				}
				errCh <- fmt.Errorf("parallel layer %q node[%d] %q failed: %w", layerName, idx, nodeName, err)
			}
		}(i, node)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
