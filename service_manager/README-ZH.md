# service_manager

`service_manager` 提供了一个用于管理服务级扩展资源的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 定位说明

`service_manager` 是一个串联其他底层模块的集成层。

- 使用 `service_manager`，意味着在服务侧把 `biz_identity`、`biz_process`、`ext_model` 串起来统一管理
- 仓库中的其他模块依然都可以单独使用
- 这些底层模块彼此之间没有强依赖关系，可以按业务需要独立接入

## 模块关系图

```text
                          +-------------------+
                          |  service_manager  |
                          |     集成管理层     |
                          +-------------------+
                            /       |       \
                           /        |        \
                          v         v         v
                +---------------+ +---------------+ +---------------+
                | biz_identity  | |  biz_process  | |   ext_model   |
                | 身份白名单管理 | | 多流程编排管理 | | 模型白名单裁剪 |
                +---------------+ +---------------+ +---------------+

独立使用关系：

  biz_identity      biz_process      ext_model      ext_spi      ext_process
       |                 |               |             |             |
       +-----------------+---------------+-------------+-------------+
                         各模块都可以独立使用
```

## 核心容器

## 核心类型

### `ServiceManager`

`ServiceManager` 是服务实例的运行时对象，负责持有初始化后的容器并管理生命周期状态。

- `Check(ctx context.Context) error`
- `Start(ctx context.Context) error`
- `Stop(ctx context.Context) error`
- `State() ServiceManagerState`
- `IdentityContainer() *IdentityContainer`
- `ProcessContainer() *ProcessContainer`
- `ModelContainer() *ModelContainer`
- `Container(name string) (any, bool)`

生命周期状态：

- `READY`
- `STARTED`
- `STOPPED`

### `ServiceManagerBuilder`

`ServiceManagerBuilder` 用于初始化标准容器、组装启动配置，并构建一个 `ServiceManager`。

- `WithIdentityScopes(scopes ...string)`
- `WithProcess(name string, process biz_process.Process)`
- `WithModelWhitelist(rpcMethod string, allowedKeys ...string)`
- `WithContainer(name string, container any)`
- `WithStartupCheck(check StartupCheck)`
- `WithLifecycle(name string, lifecycle Lifecycle)`
- `Build() (*ServiceManager, error)`

builder 默认会初始化以下标准容器：

- `identity_container`
- `process_container`
- `model_container`

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

`SPIContainer[Impl, Input, Output]` 用于按扩展定义 key 管理 `ext_spi` 实现集合。

- `Register(definition string, impl Impl) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, mode ext_spi.Mode) ([]Output, error)`

一个 `SPIContainer` 会绑定一个 `ext_spi.Template`，然后通过这个模板执行注册进去的实现。

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
	"github.com/daidai21/biz_ext_framework/ext_spi"
	"github.com/daidai21/biz_ext_framework/service_manager"
)

type userExt struct {
	key string
}

type riskSPI interface {
	Handle(ctx context.Context, input string) (string, error)
}

type riskSPIImpl struct{}

func (riskSPIImpl) Handle(ctx context.Context, input string) (string, error) {
	return "risk:" + input, nil
}

func (u userExt) Key() string {
	return u.key
}

type serverLifecycle struct{}

func (serverLifecycle) Start(ctx context.Context) error {
	return nil
}

func (serverLifecycle) Stop(ctx context.Context) error {
	return nil
}

func main() {
	spiTemplate := ext_spi.NewTemplate(func(ctx context.Context, impl riskSPI, input string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl riskSPI, input string) (string, error) {
		return impl.Handle(ctx, input)
	})
	spiContainer, err := service_manager.NewSPIContainer[riskSPI, string, string](spiTemplate)
	if err != nil {
		panic(err)
	}
	if err := spiContainer.Register("risk.audit", riskSPIImpl{}); err != nil {
		panic(err)
	}

	manager, err := service_manager.NewServiceManagerBuilder("order-service").
		WithIdentityScopes("SELLER.SHOP").
		WithProcess("order_flow", biz_process.Process{
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
		}).
		WithModelWhitelist("psm.order#CreateOrder", "user").
		WithContainer("spi_container", spiContainer).
		WithStartupCheck(func(ctx context.Context, manager *service_manager.ServiceManager) error {
			if !manager.IdentityContainer().IsAllowed("SELLER.SHOP.OPERATOR") {
				return fmt.Errorf("identity scope missing")
			}
			return nil
		}).
		WithLifecycle("http_server", serverLifecycle{}).
		Build()
	if err != nil {
		panic(err)
	}

	if err := manager.Start(context.Background()); err != nil {
		panic(err)
	}
	defer manager.Stop(context.Background())

	fmt.Println(manager.State())
	fmt.Println(manager.IdentityContainer().IsAllowed("SELLER.SHOP.OPERATOR"))
	fmt.Println(manager.ProcessContainer().Run(context.Background(), "order_flow"))

	model := ext_model.NewExtModel()
	model.Set(userExt{key: "user"})
	model.Set(userExt{key: "secret"})

	filtered, err := manager.ModelContainer().FilterForRPC("psm.order#CreateOrder", model)
	if err != nil {
		panic(err)
	}

	_, hasUser := filtered.Get("user")
	_, hasSecret := filtered.Get("secret")
	fmt.Println(hasUser, hasSecret)

	container, _ := manager.Container("spi_container")
	fmt.Println(container.(*service_manager.SPIContainer[riskSPI, string, string]).Execute(context.Background(), "risk.audit", "order", ext_spi.All))
}
```

## 开发

在模块目录下运行测试：

```bash
cd service_manager && go test ./...
```
