# biz_ctx

`biz_ctx` 提供了一个业务上下文技术组件。

该目录本身是一个独立的 Go module。

## 核心类型

### `BizSession` 与 `BizInstance`

`BizSession` 和 `BizInstance` 是用于表达关系的技术接口：

- 一个 `biz_session`
- 多个 `biz_instance`（1:N）

```go
type BizSession interface {
    BizSessionId() string
}

type BizInstance interface {
    BizInstanceId() string
}
```

### `BizCtx`

`BizCtx` 是业务上下文行为抽象：

```go
type BizCtx interface {
    Set(ctx context.Context, instance BizInstance) bool
    Get(ctx context.Context, instanceID string) (BizInstance, bool)
    Del(ctx context.Context, instanceID string) (BizInstance, bool)
    ForEach(ctx context.Context, fn func(instance BizInstance))
    List(ctx context.Context) []BizInstance
}
```

`biz_session` 直接通过 `context.Context` 传递和读取：

```go
ctx = biz_ctx.WithBizSession(ctx, session)
session, ok := biz_ctx.BizSessionFromContext(ctx)
```

## 行为说明

- `Set/Get/Del/ForEach/List` 都会从 `context.Context` 读取 `biz_session`。
- 一个会话下可以挂载多个实例。
- 同一会话内，同一个 `instanceID` 会被最新一次 `Set` 覆盖。
- 不同会话之间，即使 `instanceID` 相同也互不影响。
- `ForEach` 会基于目标会话的快照进行遍历。
- 当 `context.Context` 中没有有效 `BizSession` 时，`Set` 返回 `false`。
- 当 `context.Context` 无效时，`Get`/`Del` 返回 `false`，`List` 返回 `nil`，`ForEach` 不执行回调。

## 示例

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
