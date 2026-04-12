# biz_process

`biz_process` 提供了一个可扩展的业务流程 FSM 框架。

该目录本身是一个独立的 Go module。

## 核心类型

- `State` / `Event`：基于字符串的状态机标识。
- `Transition`：从 `From + Event` 到 `To` 的迁移规则。
- `Guard`：可选前置校验函数。
- `Action`：可选业务执行函数。
- `Extension`：迁移生命周期扩展钩子。

```go
type Transition struct {
    From   State
    Event  Event
    To     State
    Guard  Guard
    Action Action
}
```

## 扩展钩子

```go
type Extension interface {
    BeforeTransition(ctx context.Context, from State, to State, event Event, payload any) error
    AfterTransition(ctx context.Context, from State, to State, event Event, payload any)
    OnTransitionError(ctx context.Context, from State, to State, event Event, payload any, err error)
}
```

默认空实现可使用 `NoopExtension`。

## 行为说明

- `FSM` 线程安全。
- 迁移键由 `from + event` 组成。
- `Fire` 执行顺序：
1. 匹配迁移规则
2. 执行 `BeforeTransition`
3. 执行 `Guard`
4. 执行 `Action`
5. 更新状态
6. 执行 `AfterTransition`
- 在状态更新前任一步出错，状态都保持不变。
- `OnTransitionError` 会在规则缺失、钩子失败、Guard 拒绝、Action 失败时触发。



## BPMN-like 流程编排

`bpmn.go` 提供了一个轻量流程编排器，可通过 `var process = []Step{...}` 配置流程。

- 顶层 `[]Step` 按串行执行
- `Step.Parallel` 内分支并行执行
- 每个 `Step` 必须且只能配置一种模式：`Task` 或 `Parallel`

```go
process := []biz_process.Step{
    {Name: "prepare", Task: func(ctx context.Context) error { return nil }},
    {Name: "fanout", Parallel: []biz_process.Step{
        {Name: "audit", Task: func(ctx context.Context) error { return nil }},
        {Name: "notify", Task: func(ctx context.Context) error { return nil }},
    }},
    {Name: "finalize", Task: func(ctx context.Context) error { return nil }},
}

if err := biz_process.RunProcess(context.Background(), process); err != nil {
    panic(err)
}
```

## DAG 流程编排

`dag.go` 提供了基于依赖关系的 DAG 编排能力。

- 节点按依赖顺序执行
- 同一拓扑层级会并行执行
- 内置环检测与非法依赖校验

```go
dag := []biz_process.DAGNode{
    {Name: "prepare", Task: func(ctx context.Context) error { return nil }},
    {Name: "audit", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { return nil }},
    {Name: "notify", DependsOn: []string{"prepare"}, Task: func(ctx context.Context) error { return nil }},
    {Name: "finalize", DependsOn: []string{"audit", "notify"}, Task: func(ctx context.Context) error { return nil }},
}

if err := biz_process.RunDAG(context.Background(), dag); err != nil {
    panic(err)
}
```

## 示例

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
