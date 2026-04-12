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
    Set(session BizSession, instance BizInstance)
    Get(sessionID string, instanceID string) (BizInstance, bool)
    Del(sessionID string, instanceID string) (BizInstance, bool)
    ForEach(sessionID string, fn func(instance BizInstance))
    List(sessionID string) []BizInstance
}
```

## 行为说明

- `Set` 会把一个实例绑定到一个会话桶中。
- 一个会话下可以挂载多个实例。
- 同一会话内，同一个 `instanceID` 会被最新一次 `Set` 覆盖。
- 不同会话之间，即使 `instanceID` 相同也互不影响。
- `ForEach` 会基于目标会话的快照进行遍历。

## 示例

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
