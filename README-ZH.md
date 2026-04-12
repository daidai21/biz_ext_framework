# biz_ext_framework

`biz_ext_framework` 是一个用于沉淀可复用业务扩展组件的仓库。

组件按仓库顶层目录组织。其中一部分目录已经是独立 Go module，另一部分目录目前先作为占位，留待后续补充实现。

## 目录结构

- `biz_ctx/`：业务上下文组件占位目录
- `biz_identity/`：业务身份抽象的独立 Go module
- `biz_process/`：业务流程 FSM 的独立 Go module
- `ext_interceptor/`：扩展拦截器组件占位目录
- `ext_model/`：扩展模型抽象的独立 Go module
- `ext_process/`：扩展流程组件占位目录
- `ext_spi/`：SPI 模板抽象的独立 Go module
- `service_manager/`：服务管理组件占位目录
- `Makefile`：仓库级辅助命令
- `go.mod`：仓库级 Go module 定义

## 已实现模块

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

`biz_process` 提供了一个可扩展的 FSM 框架：

- `State` / `Event`
- `Transition`（`From + Event -> To`）
- `Guard`
- `Action`
- `Extension` 钩子

文档入口：

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

### `ext_spi`

`ext_spi` 提供了一个支持四种模式的通用 SPI 模板：

- `First`
- `All`
- `FirstMatched`
- `AllMatched`

文档入口：

- English: [`ext_spi/README.md`](./ext_spi/README.md)
- 中文: [`ext_spi/README-ZH.md`](./ext_spi/README-ZH.md)

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
