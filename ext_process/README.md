# ext_process

`ext_process` provides a generic extension process template.

This directory is an independent Go module.

## Core Types

### `Mode`

`ext_process` supports two execution modes:

- `Serial`: run matched implementations in registration order; stop when `continueNext=false`
- `Parallel`: run matched implementations concurrently; ignore `continueNext`

### `Template`

`Template[Impl, Input, Output]` call shape:

```go
func(ctx context.Context, extProcessImpls []Impl, input Input, mode Mode) ([]Output, error)
```

Build it with `NewTemplate(match, process)`.

### `DefinitionAction`

`ext_process` also provides definition-level merge actions for managing implementation lists under one definition:

- `Append`: append incoming implementations after the current flow
- `Skip`: ignore incoming implementations when the definition already has a flow
- `Overwrite`: replace the current flow with incoming implementations

### `AppendType`

When `DefinitionAction=Append`, `AppendType` controls where incoming implementations are inserted:

- `AppendBefore`: prepend incoming implementations before the current flow
- `AppendAfter`: append incoming implementations after the current flow
- `AppendParallel`: append incoming implementations after the current flow; typically used together with `Execute(..., ext_process.Parallel)`

### `Aspect`

If you do not use `service_manager`, business code can still bind one ext process into `context.Context` and trigger it at the beginning of a function with:

```go
defer ext_process.Aspect(ctx, input)
```

Typical usage:

```go
ctx = ext_process.BindAspect(ctx, template, extProcessImpls, ext_process.Serial)

func Handle(ctx context.Context, input OrderInput) error {
    defer ext_process.Aspect(ctx, input)
    return nil
}
```

This is useful when you want a lightweight tail-executed extension flow without depending on `service_manager`.

### `ProcessFunc`

```go
type ProcessFunc[Impl any, Input any, Output any] func(ctx context.Context, impl Impl, input Input) (output Output, continueNext bool, err error)
```

`continueNext` only affects `Serial` mode.

## Example

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

    // extProcessImpls is your ordered implementation slice.
    results, err := template(context.Background(), extProcessImpls, OrderInput{Scene: "ORDER"}, ext_process.Serial)
    if err != nil {
        panic(err)
    }

    fmt.Println(len(results))
}
```

## Development

Run tests from the module directory:

```bash
cd ext_process && go test ./...
```
