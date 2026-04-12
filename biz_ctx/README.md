# biz_ctx

`biz_ctx` provides a technical business context component.

This directory is an independent Go module.

## Core Types

### `BizSession` and `BizInstance`

`BizSession` and `BizInstance` are technical interfaces used to model the relationship:

- one `biz_session`
- many `biz_instance` (1:N)

```go
type BizSession interface {
    BizSessionId() string
}

type BizInstance interface {
    BizInstanceId() string
}
```

### `BizCtx`

`BizCtx` is the context behavior abstraction:

```go
type BizCtx interface {
    Set(ctx context.Context, instance BizInstance) bool
    Get(ctx context.Context, instanceID string) (BizInstance, bool)
    Del(ctx context.Context, instanceID string) (BizInstance, bool)
    ForEach(ctx context.Context, fn func(instance BizInstance))
    List(ctx context.Context) []BizInstance
}
```

`biz_session` is passed and read directly by `context.Context`:

```go
ctx = biz_ctx.WithBizSession(ctx, session)
session, ok := biz_ctx.BizSessionFromContext(ctx)
```

## Behavior

- `Set/Get/Del/ForEach/List` all read `biz_session` from `context.Context`.
- One session can hold many instances.
- Same `instanceID` in the same session is overwritten by latest `Set`.
- Same `instanceID` in different sessions is isolated.
- `ForEach` iterates a snapshot for the target session.
- `Set` returns `false` when `context.Context` does not carry a valid `BizSession`.
- `Get`/`Del` return `false`, `List` returns `nil`, and `ForEach` is no-op for invalid context.

## Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/daidai21/biz_ext_framework/biz_ctx"
)

type Session struct {
    id string
}

func (s Session) BizSessionId() string {
    return s.id
}

type Instance struct {
    id   string
    name string
}

func (i Instance) BizInstanceId() string {
    return i.id
}

func main() {
    bizCtx := biz_ctx.NewBizCtx()

    session := Session{id: "s1"}
    reqCtx := biz_ctx.WithBizSession(context.Background(), session)
    bizCtx.Set(reqCtx, Instance{id: "i1", name: "order"})
    bizCtx.Set(reqCtx, Instance{id: "i2", name: "refund"})

    value, ok := bizCtx.Get(reqCtx, "i1")
    if !ok {
        panic("missing instance")
    }

    instance := value.(Instance)
    fmt.Println(instance.name)
}
```
