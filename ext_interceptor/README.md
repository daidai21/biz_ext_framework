# ext_interceptor

`ext_interceptor` provides a standalone Go module for extension interceptor abstractions.

This directory is an independent Go module.

## Core Types

### `Handler`

`Handler[Input, Output]` is the final business handler wrapped by interceptors.

```go
type Handler[Input any, Output any] func(ctx context.Context, input Input) (Output, error)
```

### `Template`

`Template[Impl, Input, Output]` is the interceptor execution entry:

```go
func(ctx context.Context, interceptors []Impl, input Input, final Handler[Input, Output]) (Output, error)
```

`NewTemplate` accepts:

- `MatchFunc`
- `InterceptFunc`

and returns a template that wraps the final handler by interceptor order.

## Example

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

## Development

Run tests from the module directory:

```bash
cd ext_interceptor && go test ./...
```
