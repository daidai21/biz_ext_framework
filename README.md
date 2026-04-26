# biz_ext_framework

`biz_ext_framework` is a repository of platform components, extension components, and service-side integration utilities.

The repository is organized around small Go modules. You can either use a module independently or adopt `service_manager` as the integration layer that wires several modules together.

## 1. Highlights

- IOC-style business component container with `GlobalScope` and `SessionScope`
- business session context and identity abstractions
- lightweight observation helpers for log / metrics / trace
- process orchestration for FSM, BPMN-like layered flow, and DAG
- extension templates for SPI, process pipelines, and interceptors
- CLI tools for generating and parsing process graphs

## 2. Architecture

`service_manager` currently integrates both platform components and extension components:

- `biz_component`
- `biz_ctx`
- `biz_identity`
- `biz_observation`
- `biz_process`
- `ext_interceptor`
- `ext_model`
- `ext_process`
- `ext_spi`

All of these modules remain independently usable, while `service_manager` wires them together on the service side.

```text
                          +-------------------+
                          |  service_manager  |
                          |   integration     |
                          +-------------------+
                           /                 \
                          v                   v
               +-------------------+   +-------------------+
               |   biz_xxx modules  |   |   ext_xxx modules  |
               |platform components |   |extension components|
               +-------------------+   +-------------------+
               | biz_component     |   | ext_model         |
               | biz_ctx           |   | ext_process       |
               | biz_identity      |   | ext_spi           |
               | biz_observation   |   | ext_interceptor   |
               | biz_process       |   |                   |
               +-------------------+   +-------------------+
```

## 3. Modules

### 3.1 Platform Components

#### 3.1.1 `biz_component`

IOC-style platform component container.

Key capabilities:

- `Container`
- `GlobalScope`
- `SessionScope`
- typed `Key[T]` / `Provider[T]`
- same component name can exist in both global and session scope

Docs:

- English: [`biz_component/README.md`](./biz_component/README.md)
- 中文: [`biz_component/README-ZH.md`](./biz_component/README-ZH.md)

#### 3.1.2 `biz_process`

Platform-side process orchestration primitives.

Key capabilities:

- FSM
- BPMN-like serial-layer / parallel-node orchestration
- DAG orchestration
- standardized JSON serialization through `ProcessStringer`
- graph generation / parsing tools

Docs:

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

#### 3.1.3 Other Platform Components

- [`biz_ctx`](./biz_ctx/README.md): business session context
- [`biz_identity`](./biz_identity/README.md): business identity parsing and validation
- [`biz_observation`](./biz_observation/README.md): observation helpers

### 3.2 Extension Components

- [`ext_model`](./ext_model/README.md): extension model map abstraction
- [`ext_process`](./ext_process/README.md): extension process template
- [`ext_spi`](./ext_spi/README.md): SPI extension template
- [`ext_interceptor`](./ext_interceptor/README.md): extension interceptor template

### 3.3 Integration Layer

#### 3.3.1 `service_manager`

Service-side integration layer for container initialization, lifecycle management, observation dependencies, process orchestration, SPI registration, and model filtering.

Docs:

- English: [`service_manager/README.md`](./service_manager/README.md)
- 中文: [`service_manager/README-ZH.md`](./service_manager/README-ZH.md)

## 4. Tools

The repository also provides CLI tools under [`tools/`](./tools/README.md):

- `gen_process_graph`: generate Mermaid / DOT from BPMN, DAG, or FSM specs
- `parse_process_graph`: parse runtime metric logs and render BPMN, DAG, or FSM graphs with aggregated metrics

Install from GitHub:

```bash
go install github.com/daidai21/biz_ext_framework/tools/gen_process_graph@latest
go install github.com/daidai21/biz_ext_framework/tools/parse_process_graph@latest
```

## 5. Quick Start

Start with `service_manager` if you want one service runtime that wires platform components and extension components together:

```go
package main

import (
    "context"
    "fmt"

    "github.com/daidai21/biz_ext_framework/biz_process"
    "github.com/daidai21/biz_ext_framework/service_manager"
)

func main() {
    manager, err := service_manager.NewServiceManagerBuilder("order-service").
        WithIdentityScopes("SELLER.SHOP").
        WithProcess("order_flow", biz_process.Process{
            Layers: []biz_process.ProcessLayer{
                {
                    Name: "prepare",
                    Nodes: []biz_process.ProcessNode{
                        biz_process.Task{
                            Name: "prepare",
                            Task: func(ctx context.Context) error {
                                fmt.Println("prepare order")
                                return nil
                            },
                        },
                    },
                },
            },
        }).
        Build()
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        panic(err)
    }
    defer manager.Stop(ctx)

    fmt.Println(manager.IdentityContainer().IsAllowed("SELLER.SHOP.OPERATOR"))
    fmt.Println(manager.ProcessContainer().Run(ctx, "order_flow"))
}
```

If you only want a single module, see the README in that module directory directly.

Tool usage:

```bash
gen_process_graph -type bpmn -input process.json
parse_process_graph -type fsm -input fsm_metrics.jsonl -metrics qps,success_rate,p99
```

## 6. Repository Layout

- `biz_component/`
- `biz_ctx/`
- `biz_identity/`
- `biz_observation/`
- `biz_process/`
- `ext_interceptor/`
- `ext_model/`
- `ext_process/`
- `ext_spi/`
- `service_manager/`
- `tools/`
- `Makefile`
- `go.mod`

## 7. Development

Run tests from a target module directory:

```bash
cd biz_process && go test ./...
```

Useful repository-level commands:

```bash
make statistics_lines
make unittest
```
