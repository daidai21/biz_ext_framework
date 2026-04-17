# service_manager

`service_manager` 提供了一个用于管理服务级扩展资源的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 定位说明

`service_manager` 是一个串联其他底层模块的集成层。

- 使用 `service_manager`，意味着在服务侧把 `biz_component`、`biz_ctx`、`biz_identity`、`biz_observation`、`biz_process`、`ext_model` 以及扩展编排模块串起来统一管理
- 仓库中的其他模块依然都可以单独使用
- 这些底层模块彼此之间没有强依赖关系，可以按业务需要独立接入

## 模块关系图

```text
                          +-------------------+
                          |  service_manager  |
                          |     集成管理层     |
                          +-------------------+
                        /    /      |      \      \
                       v    v       v       v      v
                +-----------+ +-----------+ +-----------+ +-----------+ +-----------+
                |  biz_ctx  | |biz_identity| |biz_observ.| |biz_process| | ext_model |
                |session上下文| |身份白名单  | |日志/指标/链路| |多流程编排  | |模型白名单  |
                +-----------+ +-----------+ +-----------+ +-----------+ +-----------+

独立使用关系：

  biz_component  biz_ctx  biz_identity  biz_observation  biz_process  ext_model  ext_spi  ext_process  ext_interceptor
       |           |           |              |              |            |         |         |                |
       +-----------+-----------+--------------+--------------+------------+---------+---------+----------------+
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

- `component_container`
- `ctx_container`
- `identity_container`
- `observation_container`
- `process_container`
- `model_container`

### `ComponentContainer`

`ComponentContainer` 用于管理 `biz_component` 的 IOC 对象，既支持服务级，也支持 Session 级。

- `Container() *biz_component.Container`
- `RegisterAny(name string, scope biz_component.Scope, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error`
- `ResolveAny(ctx context.Context, name string) (any, error)`
- `ResolveAnyInSession(ctx context.Context, sessionID, name string) (any, error)`
- `ServiceObject(name string) (any, bool)`
- `SessionObject(sessionID, name string) (any, bool)`
- `DeleteService(name string)`
- `DeleteSessionObject(sessionID, name string)`
- `ClearSession(sessionID string)`

带类型的初始化和获取，应该通过 `Container()` 配合 `biz_component` 的泛型 helper 来完成。

### `CtxContainer`

`CtxContainer` 用于管理 `biz_ctx.BizSession` 和上下文注入。

- `Create(sessionID string) (biz_ctx.BizSession, error)`
- `Register(session biz_ctx.BizSession) error`
- `Remove(sessionID string)`
- `Get(sessionID string) (biz_ctx.BizSession, bool)`
- `SessionIDs() []string`
- `WithSession(ctx context.Context, sessionID string) (context.Context, error)`
- `SessionFromContext(ctx context.Context) (biz_ctx.BizSession, bool)`

### `ObservationContainer`

`ObservationContainer` 用于管理 `biz_observation` 依赖。

- `SetLogger(logger biz_observation.Logger)`
- `SetMetricsRecorder(recorder biz_observation.MetricsRecorder)`
- `SetTracer(tracer biz_observation.Tracer)`
- `Logger() biz_observation.Logger`
- `MetricsRecorder() biz_observation.MetricsRecorder`
- `Tracer() biz_observation.Tracer`
- `Log(...)`
- `ObserveDuration(...)`
- `StartSpan(...)`

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

### `ExtProcessContainer`

`ExtProcessContainer[Impl, Input, Output]` 用于按 definition key 管理 `ext_process` 实现集合。

- `Register(definition string, impl Impl) error`
- `RegisterWithAction(definition string, impl Impl, action ext_process.DefinitionAction) error`
- `Apply(definition string, impls []Impl, action ext_process.DefinitionAction) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, mode ext_process.Mode) ([]Output, error)`

一个 `ExtProcessContainer` 会绑定一个 `ext_process.Template`，然后通过这个模板执行注册进去的实现。

流程定义管理支持三种动作：

- `Append`：默认行为，把新实现追加到现有流程后
- `Skip`：若 definition 已存在流程，则忽略新的实现
- `Overwrite`：用新的实现列表整体覆写原流程

### `InterceptorContainer`

`InterceptorContainer[Impl, Input, Output]` 用于按 definition key 管理 `ext_interceptor` 实现集合。

- `Register(definition string, interceptor Impl) error`
- `Replace(definition string, interceptors []Impl) error`
- `Remove(definition string)`
- `Interceptors(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, final ext_interceptor.Handler[Input, Output]) (Output, error)`

一个 `InterceptorContainer` 会绑定一个 `ext_interceptor.Template`，然后通过这个模板执行注册进去的拦截器。

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
