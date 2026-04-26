# biz_ext_framework

[![Read the Docs](https://img.shields.io/badge/Read%20the%20Docs-README-8CA1AF?logo=readthedocs&logoColor=white)](./README.md)
[![Coverage](https://img.shields.io/badge/Coverage-94.2%25-brightgreen)](./unittest_coverage.md)
![Language](https://img.shields.io/badge/Language-Go-00ADD8?logo=go&logoColor=white)
![Visitors](https://visitor-badge.laobi.icu/badge?page_id=daidai21.biz_ext_framework&left_text=visitors)
![CLI Tools](https://img.shields.io/badge/CLI%20Tools-tools-6A5ACD)

![Platform Components](https://img.shields.io/badge/Platform%20Components-biz__xxx-4C8EDA)
![Extension Components](https://img.shields.io/badge/Extension%20Components-ext__xxx-F28C28)
![Service Integration](https://img.shields.io/badge/Service%20Integration-service__manager-2E8B57)

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

#### 3.1.3 `biz_ctx`

Platform-side business session context component.

Key capabilities:

- `BizSession`
- `BizSessionId`
- `WithBizSession` / `BizSessionFromContext`
- propagate and read business session state from request context

Docs:

- English: [`biz_ctx/README.md`](./biz_ctx/README.md)
- 中文: [`biz_ctx/README-ZH.md`](./biz_ctx/README-ZH.md)

#### 3.1.4 `biz_identity`

Platform-side business identity abstraction component.

Key capabilities:

- `BizIdentity`
- `ParseIdentityID`
- `ValidateIdentityID`
- `Parser` / `Validator`

Docs:

- English: [`biz_identity/README.md`](./biz_identity/README.md)
- 中文: [`biz_identity/README-ZH.md`](./biz_identity/README-ZH.md)

#### 3.1.5 `biz_observation`

Platform-side observation component.

Key capabilities:

- `Logger`
- `MetricsRecorder`
- `Tracer`
- unified abstractions and helpers for log, metrics, and tracing

Docs:

- English: [`biz_observation/README.md`](./biz_observation/README.md)
- 中文: [`biz_observation/README-ZH.md`](./biz_observation/README-ZH.md)

### 3.2 Extension Components

#### 3.2.1 `ext_model`

Extension-side model container component.

Key capabilities:

- `ExtObj`
- `ExtModel`
- `ExtMap`
- `CopyExtMap` with filter / deep-copy options

Docs:

- English: [`ext_model/README.md`](./ext_model/README.md)
- 中文: [`ext_model/README-ZH.md`](./ext_model/README-ZH.md)

#### 3.2.2 `ext_process`

Extension-side process template component.

Key capabilities:

- `Template`
- `Mode` (`Serial` / `Parallel`)
- `DefinitionAction`
- matching, merging, and execution orchestration for extension implementations

Docs:

- English: [`ext_process/README.md`](./ext_process/README.md)
- 中文: [`ext_process/README-ZH.md`](./ext_process/README-ZH.md)

#### 3.2.3 `ext_spi`

Extension-side SPI template component.

Key capabilities:

- `Template`
- `Mode` (`First` / `All` / `FirstMatched` / `AllMatched`)
- `MatchFunc`
- matching and execution template for SPI implementations

Docs:

- English: [`ext_spi/README.md`](./ext_spi/README.md)
- 中文: [`ext_spi/README-ZH.md`](./ext_spi/README-ZH.md)

#### 3.2.4 `ext_interceptor`

Extension-side interceptor template component.

Key capabilities:

- `Handler`
- `Template`
- `MatchFunc`
- matching and execution orchestration for interceptor chains

Docs:

- English: [`ext_interceptor/README.md`](./ext_interceptor/README.md)
- 中文: [`ext_interceptor/README-ZH.md`](./ext_interceptor/README-ZH.md)

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

Install dependencies first:

```bash
go get github.com/daidai21/biz_ext_framework/service_manager
go get github.com/daidai21/biz_ext_framework/biz_process
```

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
