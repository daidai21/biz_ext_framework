# biz_process

`biz_process` provides an extensible FSM framework for business process orchestration.

This directory is an independent Go module.

## Core Types

- `State` / `Event`: string-based state machine keys.
- `Transition`: transition rule from `From + Event` to `To`.
- `Guard`: optional pre-check function.
- `Node`: common lightweight node abstraction.
- `Action`: FSM node type for transition execution.
- `Extension`: extension hooks around transition lifecycle.

```go
type Transition struct {
    From   State
    Event  Event
    To     State
    Guard  Guard
    Action Action
}
```

## Extension Hooks

```go
type Extension interface {
    BeforeTransition(ctx context.Context, from State, to State, event Event, payload any) error
    AfterTransition(ctx context.Context, from State, to State, event Event, payload any)
    OnTransitionError(ctx context.Context, from State, to State, event Event, payload any, err error)
}
```

Use `NoopExtension` as a default no-op implementation.

## Behavior

- `FSM` is concurrency-safe.
- Transition key is `from + event`.
- On `Fire`:
1. resolve transition rule
2. run `BeforeTransition`
3. run `Guard`
4. run `Action`
5. update state
6. run `AfterTransition`
- Any error before state update keeps the original state.
- `OnTransitionError` is triggered for transition-not-found, hook error, guard rejection, and action error.



## BPMN-like Orchestration

`bpmn.go` provides a lightweight process orchestrator configured by `Process -> ProcessLayer -> Task`.

- `Process.Layers` runs in serial order
- `ProcessLayer.Nodes` runs `Task` nodes in parallel within the same layer
- `Task` implements both `Node` and `ProcessNode`

```go
process := biz_process.Process{
    Name: "order-flow",
    Layers: []biz_process.ProcessLayer{
        {
            Name: "prepare",
            Nodes: []biz_process.ProcessNode{
                biz_process.Task{Name: "prepare", Task: func(ctx context.Context) error { return nil }},
            },
        },
        {
            Name: "fanout",
            Nodes: []biz_process.ProcessNode{
                biz_process.Task{Name: "audit", Task: func(ctx context.Context) error { return nil }},
                biz_process.Task{Name: "notify", Task: func(ctx context.Context) error { return nil }},
            },
        },
        {
            Name: "finalize",
            Nodes: []biz_process.ProcessNode{
                biz_process.Task{Name: "finalize", Task: func(ctx context.Context) error { return nil }},
            },
        },
    },
}

if err := biz_process.RunProcess(context.Background(), process); err != nil {
    panic(err)
}
```

## DAG Orchestration

`dag.go` provides DAG-based orchestration for dependency-driven flow.

- nodes run by dependency order
- same topological level runs in parallel
- cycle / invalid dependency detection built in
- `GraphNode` implements `Node`

```go
dag := []biz_process.GraphNode{
    {Name: "prepare", Task: func(ctx context.Context) error { return nil }},
    {Name: "audit", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { return nil }},
    {Name: "notify", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { return nil }},
    {Name: "finalize", DependsOn: []string{"audit", "notify"}, Task: func(ctx context.Context) error { return nil }},
}

if err := biz_process.RunDAG(context.Background(), dag); err != nil {
    panic(err)
}
```

## Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/daidai21/biz_ext_framework/biz_process"
)

func main() {
    fsm, err := biz_process.NewFSM("CREATED", []biz_process.Transition{
        {From: "CREATED", Event: "PAY", To: "PAID"},
        {From: "PAID", Event: "SHIP", To: "SHIPPED"},
    })
    if err != nil {
        panic(err)
    }

    state, err := fsm.Fire(context.Background(), "PAY", nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(state)
}
```
