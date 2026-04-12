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
    Set(session BizSession, instance BizInstance)
    Get(sessionID string, instanceID string) (BizInstance, bool)
    Del(sessionID string, instanceID string) (BizInstance, bool)
    ForEach(sessionID string, fn func(instance BizInstance))
    List(sessionID string) []BizInstance
}
```

## Behavior

- `Set` binds one instance into one session bucket.
- One session can hold many instances.
- Same `instanceID` in the same session is overwritten by latest `Set`.
- Same `instanceID` in different sessions is isolated.
- `ForEach` iterates a snapshot for the target session.

## Example

```go
package main

import (
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
    ctx := biz_ctx.NewBizCtx()

    session := Session{id: "s1"}
    ctx.Set(session, Instance{id: "i1", name: "order"})
    ctx.Set(session, Instance{id: "i2", name: "refund"})

    value, ok := ctx.Get("s1", "i1")
    if !ok {
        panic("missing instance")
    }

    instance := value.(Instance)
    fmt.Println(instance.name)
}
```
