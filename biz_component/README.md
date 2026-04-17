# biz_component

`biz_component` provides a standalone Go module for IOC-style business component management.

This directory is an independent Go module.

## Core Concepts

- `Container`: the IOC container
- `ServiceScope`: service instance level singleton objects
- `SessionScope`: session level objects isolated by `biz_ctx.BizSessionId`
- `Provider`: lazy object constructor with dependency resolution support
- `Resolver`: dependency lookup interface used inside providers

## Behavior

- service-scope objects are created once per container instance
- session-scope objects are created once per session id
- providers can resolve other components through `Resolver`
- circular dependencies are detected
- the container is concurrency-safe

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/daidai21/biz_ext_framework/biz_component"
	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func main() {
	container := biz_component.NewContainer()

	_ = container.RegisterService("config", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		return "cfg", nil
	})
	_ = container.RegisterSession("order_component", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		cfg, err := resolver.Resolve(ctx, "config")
		if err != nil {
			return nil, err
		}
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return fmt.Sprintf("%s:%s", cfg.(string), session.BizSessionId()), nil
	})

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	value, err := container.Resolve(ctx, "order_component")
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
}
```

## Development

Run tests from the module directory:

```bash
cd biz_component && go test ./...
```
