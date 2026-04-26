# biz_component

`biz_component` provides a standalone Go module for IOC-style business component management.

This directory is an independent Go module.

## Core Concepts

- `Container`: the IOC container
- `GlobalScope`: global singleton objects
- `SessionScope`: session level objects isolated by `biz_ctx.BizSessionId`
- `Key[T]`: typed component key
- `Provider[T]`: typed lazy object constructor with dependency resolution support
- `Resolver`: dependency lookup interface used inside providers
- `Namespace`: layered component namespace

## Behavior

- global objects are created once per container instance
- session-scope objects are created once per session id
- use `GlobalScope` / `GlobalKey` / `RegisterGlobal` / `GlobalObject` for container-wide singletons
- providers resolve other components through typed generic helpers
- circular dependencies are detected
- layered namespace dependency rules are enforced
- the container is concurrency-safe

## Namespace Layers

`biz_component` defines these namespaces:

- `infra`
- `repository`
- `service`
- `domain`
- `capability`
- `business`
- `handler`

Dependency rules:

- `infra` cannot depend on anything
- `repository` cannot depend on anything
- `service` can only depend on `infra` and `repository`
- `domain` can only depend on `service`, `infra`, and `repository`
- `capability` / `business` can only depend on `domain`, `service`, `repository`, and `infra`
- `handler` can depend on all namespaces

If namespace is not specified explicitly, `GlobalKey` / `SessionKey` default to `handler`.

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
	configKey := biz_component.GlobalKeyIn[string](biz_component.ServiceNamespace, "config")
	componentKey := biz_component.SessionKeyIn[string](biz_component.HandlerNamespace, "order_component")

	_ = biz_component.RegisterGlobal(container, configKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		return "cfg", nil
	})
	_ = biz_component.RegisterSession(container, componentKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		cfg, err := biz_component.Resolve(ctx, resolver, configKey)
		if err != nil {
			return "", err
		}
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return fmt.Sprintf("%s:%s", cfg, session.BizSessionId()), nil
	})

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	value, err := biz_component.Resolve(ctx, container, componentKey)
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
