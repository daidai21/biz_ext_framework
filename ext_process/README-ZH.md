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
