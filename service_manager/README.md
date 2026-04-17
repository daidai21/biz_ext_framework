# service_manager

`service_manager` provides a standalone Go module for managing service-level extension resources.

This directory is an independent Go module.

## Positioning

`service_manager` is the integration layer that wires several lower-level modules together.

- using `service_manager` means using `biz_component`, `biz_ctx`, `biz_identity`, `biz_observation`, `biz_process`, `ext_model`, and extension orchestration modules together through one service-side management layer
- other modules in this repository can still be used independently
- those lower-level modules do not depend on each other and can be adopted separately based on business needs

## Module Diagram

```text
                          +-------------------+
                          |  service_manager  |
                          |   integration     |
                          +-------------------+
                        /    /      |      \      \
                       v    v       v       v      v
                +-----------+ +-----------+ +-----------+ +-----------+ +-----------+
                |  biz_ctx  | |biz_identity| |biz_observ.| |biz_process| | ext_model |
                |session ctx| |identity wl | |log/metric/| |multi-     | |model      |
                |           | |            | |trace      | |process     | |filter     |
                +-----------+ +-----------+ +-----------+ +-----------+ +-----------+

Independent usage:

  biz_component  biz_ctx  biz_identity  biz_observation  biz_process  ext_model  ext_spi  ext_process  ext_interceptor
       |            |          |              |              |            |         |         |                |
       +------------+----------+--------------+--------------+------------+---------+---------+----------------+
                                             each module can be used alone
```

## Core Containers

## Core Types

### `ServiceManager`

`ServiceManager` is the service instance runtime that holds initialized containers and manages lifecycle state.

- `Check(ctx context.Context) error`
- `Start(ctx context.Context) error`
- `Stop(ctx context.Context) error`
- `State() ServiceManagerState`
- `IdentityContainer() *IdentityContainer`
- `ProcessContainer() *ProcessContainer`
- `ModelContainer() *ModelContainer`
- `Container(name string) (any, bool)`

Lifecycle states:

- `READY`
- `STARTED`
- `STOPPED`

### `ServiceManagerBuilder`

`ServiceManagerBuilder` initializes standard containers, assembles startup configuration, and builds one `ServiceManager`.

- `WithIdentityScopes(scopes ...string)`
- `WithProcess(name string, process biz_process.Process)`
- `WithModelWhitelist(rpcMethod string, allowedKeys ...string)`
- `WithContainer(name string, container any)`
- `WithStartupCheck(check StartupCheck)`
- `WithLifecycle(name string, lifecycle Lifecycle)`
- `Build() (*ServiceManager, error)`

The builder creates these standard containers by default:

- `component_container`
- `ctx_container`
- `identity_container`
- `observation_container`
- `process_container`
- `model_container`

### `ComponentContainer`

`ComponentContainer` manages `biz_component` IOC objects for both service scope and session scope.

- `Container() *biz_component.Container`
- `RegisterAny(name string, scope biz_component.Scope, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error`
- `RegisterAnyIn(name string, scope biz_component.Scope, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error`
- `RegisterServiceIn(name string, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error`
- `RegisterSessionIn(name string, namespace biz_component.Namespace, provider func(ctx context.Context, resolver biz_component.Resolver) (any, error)) error`
- `ResolveAny(ctx context.Context, name string) (any, error)`
- `ResolveAnyInSession(ctx context.Context, sessionID, name string) (any, error)`
- `ServiceObject(name string) (any, bool)`
- `SessionObject(sessionID, name string) (any, bool)`
- `DeleteService(name string)`
- `DeleteSessionObject(sessionID, name string)`
- `ClearSession(sessionID string)`

If you want namespace-aware registration directly from `service_manager`, use `RegisterAnyIn`, `RegisterServiceIn`, or `RegisterSessionIn`.

Typed initialization and lookup should use the generic helpers from `biz_component` together with `Container()`.

### `CtxContainer`

`CtxContainer` manages `biz_ctx.BizSession` instances and context injection.

- `Create(sessionID string) (biz_ctx.BizSession, error)`
- `Register(session biz_ctx.BizSession) error`
- `Remove(sessionID string)`
- `Get(sessionID string) (biz_ctx.BizSession, bool)`
- `SessionIDs() []string`
- `WithSession(ctx context.Context, sessionID string) (context.Context, error)`
- `SessionFromContext(ctx context.Context) (biz_ctx.BizSession, bool)`

### `ObservationContainer`

`ObservationContainer` manages `biz_observation` dependencies.

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

`SPIContainer[Impl, Input, Output]` manages `ext_spi` implementations grouped by SPI definition key.

- `Register(definition string, impl Impl) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, mode ext_spi.Mode) ([]Output, error)`

One `SPIContainer` binds one `ext_spi.Template`, then executes registered implementations through that template.

### `ExtProcessContainer`

`ExtProcessContainer[Impl, Input, Output]` manages `ext_process` implementations grouped by definition key.

- `Register(definition string, impl Impl) error`
- `RegisterWithAction(definition string, impl Impl, action ext_process.DefinitionAction) error`
- `Apply(definition string, impls []Impl, action ext_process.DefinitionAction) error`
- `Replace(definition string, impls []Impl) error`
- `Remove(definition string)`
- `Implementations(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, mode ext_process.Mode) ([]Output, error)`

One `ExtProcessContainer` binds one `ext_process.Template`, then executes registered implementations through that template.

Definition management supports three actions:

- `Append`: default behavior, append incoming implementations after the current flow
- `Skip`: ignore incoming implementations when the definition already exists
- `Overwrite`: replace the current flow with the incoming implementation list

### `InterceptorContainer`

`InterceptorContainer[Impl, Input, Output]` manages `ext_interceptor` implementations grouped by definition key.

- `Register(definition string, interceptor Impl) error`
- `Replace(definition string, interceptors []Impl) error`
- `Remove(definition string)`
- `Interceptors(definition string) []Impl`
- `Definitions() []string`
- `Execute(ctx context.Context, definition string, input Input, final ext_interceptor.Handler[Input, Output]) (Output, error)`

One `InterceptorContainer` binds one `ext_interceptor.Template`, then executes registered interceptors through that template.

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

## Development

Run tests from the module directory:

```bash
cd service_manager && go test ./...
```
