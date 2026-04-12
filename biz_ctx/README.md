# biz_ctx

`biz_ctx` provides a business-context component where `BizInstance` is stored directly under `BizSession`.

This directory is an independent Go module.

## Core Types

### `BizInstance`

```go
type BizInstance interface {
    BizInstanceId() string
}
```

### `BizSession`

`BizSession` directly holds instances (1:N):

```go
type BizSession interface {
    BizSessionId() string
    Set(instance BizInstance)
    Get(instanceID string) (BizInstance, bool)
    Del(instanceID string) (BizInstance, bool)
    ForEach(fn func(instance BizInstance))
    List() []BizInstance
}
```

Use `NewBizSession(sessionID string)` for the default concurrency-safe implementation.

## Context Integration

`BizSession` is stored in `context.Context`:

```go
ctx = biz_ctx.WithBizSession(ctx, session)
session, ok := biz_ctx.BizSessionFromContext(ctx)
```

## Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/daidai21/biz_ext_framework/biz_ctx"
)

type Instance struct {
    id   string
    name string
}

func (i Instance) BizInstanceId() string {
    return i.id
}

func main() {
    session := biz_ctx.NewBizSession("s1")
    session.Set(Instance{id: "i1", name: "order"})
    session.Set(Instance{id: "i2", name: "refund"})

    reqCtx := biz_ctx.WithBizSession(context.Background(), session)

    fromCtx, ok := biz_ctx.BizSessionFromContext(reqCtx)
    if !ok {
        panic("missing session")
    }

    value, ok := fromCtx.Get("i1")
    if !ok {
        panic("missing instance")
    }

    instance := value.(Instance)
    fmt.Println(instance.name)
}
```
