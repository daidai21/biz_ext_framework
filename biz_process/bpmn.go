package biz_process

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrEmptyProcess = errors.New("empty process")
	ErrInvalidStep  = errors.New("invalid step")
)

// TaskFunc is one executable process step.
type TaskFunc func(ctx context.Context) error

// Step is a lightweight BPMN-like node.
// A process is configured as []Step:
// - top-level slice runs in serial order
// - step.Parallel runs in parallel as one group
// A step must use exactly one mode: Task or Parallel.
type Step struct {
	Name     string
	Task     TaskFunc
	Parallel []Step
}

// RunProcess executes process with serial+parallel orchestration.
func RunProcess(ctx context.Context, process []Step) error {
	if len(process) == 0 {
		return ErrEmptyProcess
	}

	for i, step := range process {
		if err := runStep(ctx, step); err != nil {
			return fmt.Errorf("process step[%d] %q failed: %w", i, step.Name, err)
		}
	}
	return nil
}

func runStep(ctx context.Context, step Step) error {
	hasTask := step.Task != nil
	hasParallel := len(step.Parallel) > 0

	if hasTask == hasParallel {
		return fmt.Errorf("%w: step %q must define exactly one of Task or Parallel", ErrInvalidStep, step.Name)
	}

	if hasTask {
		if err := step.Task(ctx); err != nil {
			return fmt.Errorf("task failed: %w", err)
		}
		return nil
	}

	return runParallel(ctx, step.Name, step.Parallel)
}

func runParallel(ctx context.Context, groupName string, branches []Step) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(branches))

	for i, branch := range branches {
		wg.Add(1)
		go func(idx int, step Step) {
			defer wg.Done()
			if err := runStep(ctx, step); err != nil {
				errCh <- fmt.Errorf("parallel group %q branch[%d] %q failed: %w", groupName, idx, step.Name, err)
			}
		}(i, branch)
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
