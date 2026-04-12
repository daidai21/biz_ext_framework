# biz_ctx

`biz_ctx` 提供了一个业务上下文组件：`BizInstance` 直接绑定在 `BizSession` 下。

该目录本身是一个独立的 Go module。

## 核心类型

### `BizInstance`

```go
type BizInstance interface {
    BizInstanceId() string
}
```

### `BizSession`

`BizSession` 直接维护实例集合（1:N）：

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

默认并发安全实现通过 `NewBizSession(sessionID string)` 获取。

## Context 集成

`BizSession` 通过 `context.Context` 传递：

```go
ctx = biz_ctx.WithBizSession(ctx, session)
session, ok := biz_ctx.BizSessionFromContext(ctx)
```

## 示例

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
