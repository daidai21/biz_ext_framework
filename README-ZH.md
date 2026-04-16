# biz_ext_framework

`biz_ext_framework` 是一个用于沉淀可复用业务扩展组件的仓库。

组件按仓库顶层目录组织。其中一部分目录已经是独立 Go module，另一部分目录目前先作为占位，留待后续补充实现。

## 架构说明

仓库中的模块支持两种使用方式：

- 直接单独使用底层模块，不依赖其他模块
- 使用 `service_manager` 作为服务侧集成层，把多个模块串联起来

当前 `service_manager` 集成了：

- `biz_identity`
- `biz_process`
- `ext_model`

这些底层模块彼此之间没有强依赖关系，仍然可以独立接入使用。

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

  biz_identity      biz_process      ext_model      ext_spi      ext_process      ext_interceptor
       |                 |               |             |             |                   |
       +-----------------+---------------+-------------+-------------+-------------------+
                                       各模块都可以独立使用
```

## 目录结构

- `biz_ctx/`：业务上下文组件占位目录
- `biz_identity/`：业务身份抽象的独立 Go module
- `biz_process/`：业务流程 FSM 的独立 Go module
- `ext_interceptor/`：扩展拦截器抽象的独立 Go module
- `ext_model/`：扩展模型抽象的独立 Go module
- `ext_process/`：扩展流程模板的独立 Go module
- `ext_spi/`：SPI 模板抽象的独立 Go module
- `service_manager/`：服务侧集成与容器管理的独立 Go module
- `Makefile`：仓库级辅助命令
- `go.mod`：仓库级 Go module 定义

## 已实现模块

### `service_manager`

`service_manager` 提供了一个构建在其他复用模块之上的服务侧集成层：

- `ServiceManager`：服务实例生命周期管理
- `ServiceManagerBuilder`：容器初始化与服务构建
- `IdentityContainer`：业务身份白名单管理
- `ProcessContainer`：多个具名流程编排管理
- `SPIContainer`：扩展定义到扩展实现的管理
- `InterceptorContainer`：拦截器定义到拦截器实现的管理
- `ModelContainer`：外调 RPC 前的 ext model 白名单裁剪

文档入口：

- English: [`service_manager/README.md`](./service_manager/README.md)
- 中文: [`service_manager/README-ZH.md`](./service_manager/README-ZH.md)

### `ext_model`

`ext_model` 提供了一个泛型、并发安全的模型 Map 抽象：

- `ExtObj`：值对象约束，要求实现 `Key() string`
- `ExtModel[V]`：Map 行为接口
- `ExtMap[V]`：默认实现
- `CopyExtMap`：复制工具，支持 `WithDeepCopy` 和 `WithKeyFilter`

文档入口：

- English: [`ext_model/README.md`](./ext_model/README.md)
- 中文: [`ext_model/README-ZH.md`](./ext_model/README-ZH.md)

### `biz_identity`

`biz_identity` 提供了一个面向业务身份抽象的技术组件：

- `BizIdentity`
- `Parser`
- `Validator`

文档入口：

- English: [`biz_identity/README.md`](./biz_identity/README.md)
- 中文: [`biz_identity/README-ZH.md`](./biz_identity/README-ZH.md)

### `biz_process`

`biz_process` 提供了流程编排相关组件：

- FSM
- BPMN-like 串行层 / 并行节点编排
- DAG 编排

文档入口：

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

### `ext_process`

`ext_process` 提供了一个通用扩展处理模板：

- `Mode`（`Serial`、`Parallel`）
- `Template`
- `MatchFunc`
- `ProcessFunc`（串行模式支持 `continueNext` 控制后续执行）

文档入口：

- English: [`ext_process/README.md`](./ext_process/README.md)
- 中文: [`ext_process/README-ZH.md`](./ext_process/README-ZH.md)

### `ext_spi`

`ext_spi` 提供了一个支持四种模式的通用 SPI 模板：

- `First`
- `All`
- `FirstMatched`
- `AllMatched`

文档入口：

- English: [`ext_spi/README.md`](./ext_spi/README.md)
- 中文: [`ext_spi/README-ZH.md`](./ext_spi/README-ZH.md)

### `ext_interceptor`

`ext_interceptor` 提供了一个通用拦截器模板抽象：

- `Handler`
- `Template`
- `MatchFunc`
- `InterceptFunc`

文档入口：

- English: [`ext_interceptor/README.md`](./ext_interceptor/README.md)
- 中文: [`ext_interceptor/README-ZH.md`](./ext_interceptor/README-ZH.md)

## 快速开始

```go
package main

import (
    "fmt"

    "github.com/daidai21/biz_ext_framework/ext_model"
)

type User struct {
    ID   string
    Name string
}

func (u User) Key() string {
    return u.ID
}

func main() {
    var users ext_model.ExtModel[User] = &ext_model.ExtMap[User]{}

    users.Set(User{ID: "u1", Name: "Alice"})

    user, ok := users.Get("u1")
    fmt.Println(user.Name, ok)
}
```

## 开发

在目标模块目录下运行测试：

```bash
cd ext_model && go test ./...
```

仓库级辅助命令：

```bash
make statistics_lines
```
