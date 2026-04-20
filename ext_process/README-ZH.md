# ext_process

`ext_process` 提供了一个通用的扩展处理模板。

该目录本身是一个独立的 Go module。

## 核心类型

### `Mode`

`ext_process` 支持两种执行模式：

- `Serial`：按注册顺序串行执行；当 `continueNext=false` 时终止后续
- `Parallel`：并行执行；忽略 `continueNext`

### `Template`

`Template[Impl, Input, Output]` 调用形态：

```go
func(ctx context.Context, extProcessImpls []Impl, input Input, mode Mode) ([]Output, error)
```

通过 `NewTemplate(match, process)` 构造。

### `DefinitionAction`

`ext_process` 也提供了定义级别的合并动作，用于管理同一个 definition 下的实现列表：

- `Append`：把新实现追加到现有流程后面
- `Skip`：如果该 definition 已经存在流程，则跳过本次新增
- `Overwrite`：使用新实现覆写已有流程

### `AppendType`

当 `DefinitionAction=Append` 时，`AppendType` 用来控制新实现插入到流程的哪个位置：

- `AppendBefore`：前插到现有流程之前
- `AppendAfter`：后插到现有流程之后
- `AppendParallel`：按后插合并，通常配合 `Execute(..., ext_process.Parallel)` 使用

### `Aspect`

如果不使用 `service_manager`，业务侧也可以直接把扩展流程绑定到 `context.Context`，然后在函数开始处调用：

```go
defer ext_process.Aspect(ctx, input)
```

常见写法如下：

```go
ctx = ext_process.BindAspect(ctx, template, extProcessImpls, ext_process.Serial)

func Handle(ctx context.Context, input OrderInput) error {
    defer ext_process.Aspect(ctx, input)
    return nil
}
```

适合把一些统一的补充流程挂到业务函数尾部执行，而不依赖 `service_manager` 的容器管理。

### `ProcessFunc`

```go
type ProcessFunc[Impl any, Input any, Output any] func(ctx context.Context, impl Impl, input Input) (output Output, continueNext bool, err error)
```

`continueNext` 只在 `Serial` 模式生效。

## 示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/daidai21/biz_ext_framework/ext_process"
)

type OrderInput struct {
    Scene string
}

type OrderProcess interface {
    Match(ctx context.Context, input OrderInput) (bool, error)
    Handle(ctx context.Context, input OrderInput) (string, bool, error)
}

func main() {
    template := ext_process.NewTemplate(
        func(ctx context.Context, impl OrderProcess, input OrderInput) (bool, error) {
            return impl.Match(ctx, input)
        },
        func(ctx context.Context, impl OrderProcess, input OrderInput) (string, bool, error) {
            return impl.Handle(ctx, input)
        },
    )

    // extProcessImpls 是你注册好的实现列表。
    results, err := template(context.Background(), extProcessImpls, OrderInput{Scene: "ORDER"}, ext_process.Serial)
    if err != nil {
        panic(err)
    }

    fmt.Println(len(results))
}
```

## 开发

在模块目录下执行测试：

```bash
cd ext_process && go test ./...
```
