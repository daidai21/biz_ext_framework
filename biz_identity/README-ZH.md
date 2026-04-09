# biz_identity

`biz_identity` 提供了一个面向业务身份抽象的技术组件。

该目录本身是一个独立的 Go module。

## 核心类型

### `BizIdentity`

`BizIdentity` 只是一个技术组件接口，不承载具体业务字段和业务行为定义：

```go
type BizIdentity interface {
    IdentityId() string
    Priority() int
}
```

### `Parser` 与 `Validator`

组件同时提供业务身份解析、校验接口：

```go
type Parser[T BizIdentity] interface {
    Parser(info map[string]string) (T, error)
}

type Validator[T BizIdentity] interface {
    Validate(identity T) error
}
```

同时提供函数适配器：

- `ParseFunc`
- `ValidateFunc`
- `DefaultValidator`

## 默认校验规则

`DefaultValidator` 会对 `IdentityId()` 执行默认格式校验：

- 身份层级数量可以是 1 到 10 级
- 每一级之间使用 `.`
- 每一级只能包含大写字母
- 开头和结尾都必须是大写字母

合法示例：

```text
SELLER
SELLER.SHOP
SELLER.SHOP.OPERATOR
```

## 示例

```go
package main

import (
    "fmt"

    "github.com/daidai21/biz_ext_framework/biz_identity"
)

type SellerIdentity struct {
    id       string
    priority int
}

func (s SellerIdentity) IdentityId() string {
    return s.id
}

func (s SellerIdentity) Priority() int {
    return s.priority
}

func main() {
    parser := biz_identity.ParseFunc[SellerIdentity](func(info map[string]string) (SellerIdentity, error) {
        return SellerIdentity{
            id:       info["identity_id"],
            priority: 10,
        }, nil
    })
    validator := biz_identity.DefaultValidator[SellerIdentity]{}

    identity, err := parser.Parser(map[string]string{
        "identity_id": "SELLER.SHOP.OPERATOR",
    })
    if err != nil {
        panic(err)
    }
    if err := validator.Validate(identity); err != nil {
        panic(err)
    }

    fmt.Println(identity.IdentityId(), identity.Priority())
}
```
