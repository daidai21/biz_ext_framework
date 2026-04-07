# biz_ext_framework

`biz_ext_framework` 是一个用于沉淀可复用业务扩展组件的 Go 仓库。

## 包概览

### `ext_model`

`ext_model` 提供了一个泛型、并发安全的模型 Map 抽象：

- `ExtObj`：值对象约束，要求实现 `Key() string`
- `ExtModel[V]`：Map 行为接口
- `ExtMap[V]`：默认实现
- `CopyExtMap`：复制工具，支持 `WithDeepCopy` 和 `WithKeyFilter`

文档入口：

- English: [`ext_model/README.md`](./ext_model/README.md)
- 中文: [`ext_model/README-ZH.md`](./ext_model/README-ZH.md)

## 快速开始

```go
package main

import (
    "fmt"

    "biz_ext_framework/ext_model"
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

运行测试：

```bash
go test ./...
```
