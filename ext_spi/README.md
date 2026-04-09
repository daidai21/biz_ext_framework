# ext_spi

`ext_spi` provides a standalone Go module for extension SPI abstractions.

This directory is an independent Go module.

## Core Types

### `Mode`

`ext_spi` supports four SPI template modes:

- `First`: invoke only the first implementation
- `All`: invoke every implementation
- `FirstMatched`: invoke only the first implementation that passes `match`
- `AllMatched`: invoke every implementation that passes `match`

### `Template`

`Template[Impl, Input, Output]` is a SPI template function. Its call shape is:

```go
func(ctx context.Context, extSpiImpls []Impl, input Input, mode Mode) ([]Output, error)
```

`NewTemplate` accepts `MatchFunc` and `InvokeFunc`, and returns that template function.

SPI implementations are expected to be methods on empty structs, and different implementations are usually assembled as the same business interface slice.

## Example

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

## Development

Run tests from the module directory:

```bash
cd ext_spi && go test ./...
```
