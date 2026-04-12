# ext_model

`ext_model` 提供了一个面向业务模型的轻量级泛型 Map 容器。

该目录本身是一个独立的 Go module。

## 核心类型

### `ExtObj`

所有存入 `ExtMap` 的值都必须实现下面这个接口：

```go
type ExtObj interface {
    Key() string
}
```

`Key()` 是对象在 `ExtMap` 中的唯一键来源。

### `ExtModel[V ExtObj]`

`ExtModel` 是 `ExtMap` 对外暴露的行为接口。

```go
type ExtModel[V ExtObj] interface {
    Get(key string) (V, bool)
    Set(value V)
    Del(key string) (V, bool)
    ForEach(fn func(value V))
}
```

### `ExtMap[V ExtObj]`

`ExtMap` 是 `ExtModel` 的并发安全实现，底层基于 `map[string]V`。

公开方法：

- `Set(value V)`
- `Get(key string) (V, bool)`
- `Del(key string) (V, bool)`
- `ForEach(fn func(value V))`

## 行为说明

- `ExtMap` 的零值可直接使用。
- `Set` 始终通过 `value.Key()` 写入。
- `Get` 和 `Del` 通过显式字符串 key 操作。
- `ForEach` 会遍历当前 Map 中的值。
- 遍历顺序不保证稳定。

## 示例

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
    users.Set(User{ID: "u2", Name: "Bob"})

    user, ok := users.Get("u1")
    fmt.Println(user.Name, ok)

    users.ForEach(func(value User) {
        fmt.Println(value.ID, value.Name)
    })

    users.Del("u2")
}
```

## 使用教程：在 `UserDO` 上挂载多个扩展结构

下面示例演示如何在一个用户对象上挂载多个扩展信息。  
注意：`ExtModel` 是接口类型，业务字段建议直接使用 `ext_model.ExtModel[...]`（而不是 `*ext_model.ExtModel`）。

```go
package main

import (
	"github.com/daidai21/biz_ext_framework/ext_model"
	"testing"
)

type userInfo struct {
	UserId int64
	ext_model.ExtModel
}

var (
	_ ext_model.ExtObj = userTaxInfo{}
	_ ext_model.ExtObj = userPhdInfo{}
)

type userTaxInfo struct {
	TaxId string
}

func (u userTaxInfo) Key() string {
	return "userTaxInfo"
}

type userPhdInfo struct {
	PhdId string
}

func (u userPhdInfo) Key() string {
	return "userPhdInfo"
}

func main(t *testing.T) {
	info := userInfo{
		UserId:   1,
		ExtModel: ext_model.NewExtModel(),
	}
	info.Set(userTaxInfo{TaxId: "tax_2313"})
	info.Set(userPhdInfo{PhdId: "phd_6748392"})
	t.Log(info)
	info.ForEach(func(value ExtObj) {
		t.Log(value)
	})
}

```

## Copy 工具

如果需要复制一个 `ExtMap`，可以使用 `ext_model.CopyExtMap`：

```go
import "github.com/daidai21/biz_ext_framework/ext_model"

copied := ext_model.CopyExtMap(src)
filtered := ext_model.CopyExtMap(src, ext_model.WithKeyFilter[User](func(key string) bool {
    return key == "u1"
}))
deepCopied := ext_model.CopyExtMap(src, ext_model.WithDeepCopy[User](func(value User) User {
    return value
}))
```

可用选项：

- `WithDeepCopy`
- `WithKeyFilter`
