# ext_spi

`ext_spi` 提供了一个用于扩展 SPI 抽象的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心类型

### `Mode`

`ext_spi` 当前支持四种 SPI 模板模式：

- `First`：只执行第一个实现
- `All`：执行全部实现
- `FirstMatched`：只执行首个通过 `match` 判断的实现
- `AllMatched`：执行全部通过 `match` 判断的实现

### `Template`

`Template[Impl, Input, Output]` 是一个 SPI 模板函数，调用形式是：

```go
func(ctx context.Context, extSpiImpls []Impl, input Input, mode Mode) ([]Output, error)
```

`NewTemplate` 接收 `MatchFunc` 和 `InvokeFunc`，并返回这个模板函数。

SPI 实现约定绑定在空结构体方法上，不同实现通常以同一个业务接口切片的形式传入。

## 示例

```go
package main

import (
	"context"
	"fmt"

	"github.com/daidai21/biz_ext_framework/ext_spi"
)

type OrderInput struct {
	Scene string
}

type OrderSPI interface {
	Match(ctx context.Context, input OrderInput) (bool, error)
	Handle(ctx context.Context, input OrderInput) (string, error)
}

type PrimarySPI struct{}

func (PrimarySPI) Match(_ context.Context, input OrderInput) (bool, error) {
	if input.Scene != "ORDER" {
		return false, nil
	}
	return true, nil
}

func (PrimarySPI) Handle(_ context.Context, input OrderInput) (string, error) {
	return "primary", nil
}

type SecondarySPI struct{}

func (SecondarySPI) Match(_ context.Context, input OrderInput) (bool, error) {
	if input.Scene != "ORDER" {
		return false, nil
	}
	return true, nil
}

func (SecondarySPI) Handle(_ context.Context, input OrderInput) (string, error) {
	return "secondary", nil
}

func main() {
	template := ext_spi.NewTemplate(func(ctx context.Context, impl OrderSPI, input OrderInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl OrderSPI, input OrderInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []OrderSPI{
		PrimarySPI{},
		SecondarySPI{},
	}, OrderInput{Scene: "ORDER"}, ext_spi.AllMatched)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(results), results[0], results[1])
}
```

## 开发

在模块目录下运行测试：

```bash
cd ext_spi && go test ./...
```
