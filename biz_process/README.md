# biz_process

`biz_process` provides an extensible FSM framework for business process orchestration.

This directory is an independent Go module.

## Core Types

- `State` / `Event`: string-based state machine keys.
- `Transition`: transition rule from `From + Event` to `To`.
- `Guard`: optional pre-check function.
- `Action`: optional business action function.
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
