package biz_process

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrEmptyDAG   = errors.New("empty dag")
	ErrInvalidDAG = errors.New("invalid dag")
	ErrDAGCycle   = errors.New("dag contains cycle")
)

type GraphTask func(ctx context.Context) error

// DAGTask is kept as a compatibility alias.
type DAGTask = GraphTask

// GraphNode describes one node in a DAG process.
type GraphNode struct {
	Name      string
	DependsOn []string
	Task      GraphTask
}

// DAGNode is kept as a compatibility alias.
type DAGNode = GraphNode

func (n GraphNode) NodeName() string {
	return n.Name
}

// RunDAG executes DAG nodes by dependency order.
// Nodes in the same topological level run in parallel.
func RunDAG(ctx context.Context, nodes []GraphNode) error {
	if len(nodes) == 0 {
		return ErrEmptyDAG
	}

	index := make(map[string]GraphNode, len(nodes))
	indegree := make(map[string]int, len(nodes))
	dependents := make(map[string][]string, len(nodes))

	for i, node := range nodes {
		if node.Name == "" {
			return fmt.Errorf("%w: node[%d] name is required", ErrInvalidDAG, i)
		}
		if node.Task == nil {
			return fmt.Errorf("%w: node %q task is required", ErrInvalidDAG, node.Name)
		}
		if _, exists := index[node.Name]; exists {
			return fmt.Errorf("%w: duplicate node name %q", ErrInvalidDAG, node.Name)
		}
		index[node.Name] = node
		indegree[node.Name] = 0
	}

	for _, node := range nodes {
		for _, dep := range node.DependsOn {
			if dep == node.Name {
				return fmt.Errorf("%w: node %q cannot depend on itself", ErrInvalidDAG, node.Name)
			}
			if _, exists := index[dep]; !exists {
				return fmt.Errorf("%w: node %q depends on unknown node %q", ErrInvalidDAG, node.Name, dep)
			}
			indegree[node.Name]++
			dependents[dep] = append(dependents[dep], node.Name)
		}
	}

	ready := make([]string, 0, len(nodes))
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}

	executed := 0
	for len(ready) > 0 {
		current := ready
		ready = nil

		errCh := make(chan error, len(current))
		var wg sync.WaitGroup
		for _, name := range current {
			wg.Add(1)
			node := index[name]
			go func(nodeName string, n GraphNode) {
				defer wg.Done()
				if err := n.Task(ctx); err != nil {
					errCh <- fmt.Errorf("dag node %q failed: %w", nodeName, err)
				}
			}(name, node)
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

		executed += len(current)
		for _, name := range current {
			for _, next := range dependents[name] {
				indegree[next]--
				if indegree[next] == 0 {
					ready = append(ready, next)
				}
			}
		}
	}

	if executed != len(nodes) {
		return ErrDAGCycle
	}
	return nil
}
