# biz_ext_framework

`biz_ext_framework` 是一个沉淀平台组件、扩展组件和服务侧集成能力的仓库。

仓库以多个小型 Go module 组织。你可以单独使用某个模块，也可以使用 `service_manager` 作为集成层，把多个模块组合到同一个服务运行时里。

## 1. 亮点

- 支持 `GlobalScope` / `SessionScope` 的 IOC 业务组件容器
- 业务 session 上下文与业务身份抽象
- 轻量日志 / 指标 / 链路观测工具
- FSM、BPMN-like、DAG 三类流程编排
- SPI、流程模板、拦截器等扩展模板
- 面向流程图生成和运行时打点解析的 CLI 工具

## 2. 架构

当前 `service_manager` 已集成平台组件与扩展组件：

- `biz_component`
- `biz_ctx`
- `biz_identity`
- `biz_observation`
- `biz_process`
- `ext_interceptor`
- `ext_model`
- `ext_process`
- `ext_spi`

这些模块都可以独立使用，`service_manager` 负责在服务侧把它们统一串起来。

```text
                          +-------------------+
                          |  service_manager  |
                          |     集成管理层     |
                          +-------------------+
                           /                 \
                          v                   v
               +-------------------+   +-------------------+
               |   biz_xxx modules  |   |   ext_xxx modules  |
               |     平台组件        |   |      扩展组件       |
               +-------------------+   +-------------------+
               | biz_component     |   | ext_model         |
               | biz_ctx           |   | ext_process       |
               | biz_identity      |   | ext_spi           |
               | biz_observation   |   | ext_interceptor   |
               | biz_process       |   |                   |
               +-------------------+   +-------------------+
```

## 3. 模块

### 3.1 平台组件

#### 3.1.1 `biz_component`

IOC 风格平台组件容器。

核心能力：

- `Container`
- `GlobalScope`
- `SessionScope`
- 泛型 `Key[T]` / `Provider[T]`
- 同一个组件名可同时存在 `Global` 与 `Session` 两个作用域

文档：

- English: [`biz_component/README.md`](./biz_component/README.md)
- 中文: [`biz_component/README-ZH.md`](./biz_component/README-ZH.md)

#### 3.1.2 `biz_process`

平台侧流程编排能力。

核心能力：

- FSM
- BPMN-like 串行层 / 并行节点编排
- DAG 编排
- 通过 `ProcessStringer` 提供标准化 JSON 序列化
- 流程图生成与运行时打点解析工具

文档：

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

#### 3.1.3 其他平台组件

- [`biz_ctx`](./biz_ctx/README.md)：业务 session 上下文
- [`biz_identity`](./biz_identity/README.md)：业务身份解析与校验
- [`biz_observation`](./biz_observation/README.md)：观测工具

### 3.2 扩展组件

- [`ext_model`](./ext_model/README.md)：扩展模型 Map
- [`ext_process`](./ext_process/README.md)：扩展流程模板
- [`ext_spi`](./ext_spi/README.md)：SPI 扩展模板
- [`ext_interceptor`](./ext_interceptor/README.md)：扩展拦截器模板

### 3.3 集成层

#### 3.3.1 `service_manager`

服务侧集成层，负责容器初始化、生命周期管理、观测依赖注入、流程编排、SPI 注册和模型过滤。

文档：

- English: [`service_manager/README.md`](./service_manager/README.md)
- 中文: [`service_manager/README-ZH.md`](./service_manager/README-ZH.md)

## 4. Tools

仓库在 [`tools/`](./tools/README.md) 下提供了 CLI 工具：

- `gen_process_graph`：从 BPMN、DAG、FSM 规格生成 Mermaid / DOT
- `parse_process_graph`：从运行时打点日志解析 BPMN、DAG、FSM 图，并聚合指标

支持从 GitHub 直接安装：

```bash
go install github.com/daidai21/biz_ext_framework/tools/gen_process_graph@latest
go install github.com/daidai21/biz_ext_framework/tools/parse_process_graph@latest
```

## 5. 快速开始

推荐从 `service_manager` 开始，把平台组件和扩展组件统一接入到一个服务运行时里：

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

如果只想单独使用某个模块，直接看对应子目录下的 README 即可。

工具使用：

```bash
gen_process_graph -type bpmn -input process.json
parse_process_graph -type fsm -input fsm_metrics.jsonl -metrics qps,success_rate,p99
```

## 6. 仓库结构

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

## 7. 开发

在目标模块目录下运行测试：

```bash
cd biz_process && go test ./...
```

常用仓库级命令：

```bash
make statistics_lines
make unittest
```
