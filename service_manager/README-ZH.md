# service_manager

`service_manager` 提供了一个用于管理服务级扩展资源的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心容器

### `IdentityContainer`

`IdentityContainer` 用于管理业务身份定义范围的白名单。

- `AllowScope(scope string) error`
- `RevokeScope(scope string)`
- `IsAllowed(identityID string) bool`
- `IsIdentityAllowed(identity biz_identity.BizIdentity) bool`
- `Scopes() []string`

白名单匹配基于业务身份层级前缀。

示例：

- `SELLER.SHOP` 可以匹配 `SELLER.SHOP`
- `SELLER.SHOP` 可以匹配 `SELLER.SHOP.OPERATOR`
- `SELLER.SHOP` 不会匹配 `SELLER.CENTER`

### `ProcessContainer`

`ProcessContainer` 用于管理多个具名的 `biz_process.Process`。

- `Register(name string, process biz_process.Process) error`
- `Unregister(name string)`
- `Get(name string) (biz_process.Process, bool)`
- `Names() []string`
- `Run(ctx context.Context, name string) error`

它适合在服务侧统一注册和按名称执行不同的流程编排定义。

### `SPIContainer`

`SPIContainer[Impl]` 用于按扩展定义 key 管理一组扩展实现。

- `Register(definition string, impl Impl) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`

返回实现列表时会保持注册顺序。

### `ModelContainer`

`ModelContainer` 用于管理外调 RPC 前的 ext model 白名单策略。

- `SetWhitelist(rpcMethod string, allowedKeys []string) error`
- `RemoveWhitelist(rpcMethod string)`
- `Whitelist(rpcMethod string) []string`
- `FilterForRPC(rpcMethod string, src ext_model.ExtModel) (ext_model.ExtModel, error)`

RPC 方法标识格式为 `PSM#Method`。

`FilterForRPC` 会返回一个复制后的 `ext_model.ExtModel`，只保留显式配置在白名单中的 key。如果某个 RPC 方法没有配置白名单，则默认返回空模型。

## 示例

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

## 开发

在模块目录下运行测试：

```bash
cd service_manager && go test ./...
```
