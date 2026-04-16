# service_manager

`service_manager` provides a standalone Go module for managing service-level extension resources.

This directory is an independent Go module.

## Positioning

`service_manager` is the integration layer that wires several lower-level modules together.

- using `service_manager` means using `biz_identity`, `biz_process`, and `ext_model` together through one service-side management layer
- other modules in this repository can still be used independently
- those lower-level modules do not depend on each other and can be adopted separately based on business needs

## Module Diagram

```text
                          +-------------------+
                          |  service_manager  |
                          |   integration     |
                          +-------------------+
                            /       |       \
                           /        |        \
                          v         v         v
                +---------------+ +---------------+ +---------------+
                | biz_identity  | |  biz_process  | |   ext_model   |
                | identity wl   | | multi-process | | model filter  |
                +---------------+ +---------------+ +---------------+

Independent usage:

  biz_identity      biz_process      ext_model      ext_spi      ext_process
       |                 |               |             |             |
       +-----------------+---------------+-------------+-------------+
                         each module can be used alone
```

## Core Containers

### `IdentityContainer`

`IdentityContainer` manages whitelisted business identity scopes.

- `AllowScope(scope string) error`
- `RevokeScope(scope string)`
- `IsAllowed(identityID string) bool`
- `IsIdentityAllowed(identity biz_identity.BizIdentity) bool`
- `Scopes() []string`

Scope matching is prefix-based on business identity levels.

Example:

- `SELLER.SHOP` matches `SELLER.SHOP`
- `SELLER.SHOP` matches `SELLER.SHOP.OPERATOR`
- `SELLER.SHOP` does not match `SELLER.CENTER`

### `ProcessContainer`

`ProcessContainer` manages multiple named `biz_process.Process` values.

- `Register(name string, process biz_process.Process) error`
- `Unregister(name string)`
- `Get(name string) (biz_process.Process, bool)`
- `Names() []string`
- `Run(ctx context.Context, name string) error`

It is intended for service-side orchestration registration, so different process definitions can be assembled and invoked by name.

### `SPIContainer`

`SPIContainer[Impl]` manages extension implementations grouped by SPI definition key.

- `Register(definition string, impl Impl) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`

Implementations are returned in registration order.

### `ModelContainer`

`ModelContainer` manages ext model whitelist policies for outbound RPC calls.

- `SetWhitelist(rpcMethod string, allowedKeys []string) error`
- `RemoveWhitelist(rpcMethod string)`
- `Whitelist(rpcMethod string) []string`
- `FilterForRPC(rpcMethod string, src ext_model.ExtModel) (ext_model.ExtModel, error)`

The RPC method key format is `PSM#Method`.

`FilterForRPC` returns a copied `ext_model.ExtModel` that only keeps explicitly whitelisted keys. If a RPC method has no configured whitelist, it returns an empty model by default.

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/daidai21/biz_ext_framework/biz_process"
	"github.com/daidai21/biz_ext_framework/ext_model"
	"github.com/daidai21/biz_ext_framework/service_manager"
)

type userExt struct {
	key string
}

func (u userExt) Key() string {
	return u.key
}

func main() {
	identityContainer, err := service_manager.NewIdentityContainer("SELLER.SHOP")
	if err != nil {
		panic(err)
	}
	fmt.Println(identityContainer.IsAllowed("SELLER.SHOP.OPERATOR"))

	processContainer := service_manager.NewProcessContainer()
	err = processContainer.Register("order_flow", biz_process.Process{
		Layers: []biz_process.ProcessLayer{
			{
				Name: "prepare",
				Nodes: []biz_process.ProcessNode{
					biz_process.TaskProcessNode{
						Name: "prepare",
						Task: func(ctx context.Context) error { return nil },
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	if err := processContainer.Run(context.Background(), "order_flow"); err != nil {
		panic(err)
	}

	spiContainer := service_manager.NewSPIContainer[string]()
	if err := spiContainer.Register("risk.audit", "impl-a"); err != nil {
		panic(err)
	}
	fmt.Println(spiContainer.Implementations("risk.audit"))

	modelContainer := service_manager.NewModelContainer()
	if err := modelContainer.SetWhitelist("psm.order#CreateOrder", []string{"user"}); err != nil {
		panic(err)
	}

	model := ext_model.NewExtModel()
	model.Set(userExt{key: "user"})
	model.Set(userExt{key: "secret"})

	filtered, err := modelContainer.FilterForRPC("psm.order#CreateOrder", model)
	if err != nil {
		panic(err)
	}
	_, hasUser := filtered.Get("user")
	_, hasSecret := filtered.Get("secret")
	fmt.Println(hasUser, hasSecret)
}
```

## Development

Run tests from the module directory:

```bash
cd service_manager && go test ./...
```
