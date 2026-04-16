# ext_interceptor

`ext_interceptor` 提供了一个用于扩展拦截器抽象的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心类型

### `Handler`

`Handler[Input, Output]` 表示被拦截器包裹的最终业务处理函数。

```go
type Handler[Input any, Output any] func(ctx context.Context, input Input) (Output, error)
```

### `Template`

`Template[Impl, Input, Output]` 是拦截器执行入口：

```go
func(ctx context.Context, interceptors []Impl, input Input, final Handler[Input, Output]) (Output, error)
```

`NewTemplate` 接收：

- `MatchFunc`
- `InterceptFunc`

并返回一个按拦截器顺序包裹最终处理函数的模板。

## 示例

```go
package main

import (
	"context"
	"fmt"

	"github.com/daidai21/biz_ext_framework/ext_interceptor"
)

type LoggingInterceptor interface {
	Handle(ctx context.Context, input string, next ext_interceptor.Handler[string, string]) (string, error)
}

type LoggingImpl struct{}

func (LoggingImpl) Handle(ctx context.Context, input string, next ext_interceptor.Handler[string, string]) (string, error) {
	output, err := next(ctx, input+"->log")
	if err != nil {
		return "", err
	}
	return output + "->after-log", nil
}

func main() {
	template := ext_interceptor.NewTemplate(func(ctx context.Context, impl LoggingInterceptor, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl LoggingInterceptor, input string, next ext_interceptor.Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})

	output, err := template(context.Background(), []LoggingInterceptor{LoggingImpl{}}, "start", func(ctx context.Context, input string) (string, error) {
		return input + "->final", nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}
```

## 开发

在模块目录下运行测试：

```bash
cd ext_interceptor && go test ./...
```
