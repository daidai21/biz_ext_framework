# biz_identity

`biz_identity` provides a technical component for business identity abstraction.

This directory is an independent Go module.

## Core Types

### `BizIdentity`

`BizIdentity` is only a technical interface. It does not define business fields or business behavior.

```go
type BizIdentity interface {
    IdentityId() string
    Priority() int
}
```

### `Parser` and `Validator`

The package also provides business identity parsing and validation interfaces:

```go
type Parser[T BizIdentity] interface {
    Parser(info map[string]string) (T, error)
}

type Validator[T BizIdentity] interface {
    Validate(identity T) error
}
```

Function adapters are also available:

- `ParseFunc`
- `ValidateFunc`
- `DefaultValidator`

## Default Validation

`DefaultValidator` validates `IdentityId()` with these rules:

- it may contain 1 to 10 levels
- levels must be separated by `.`
- each level may only contain uppercase letters
- the first and last characters must be uppercase letters

Valid example:

```text
SELLER
SELLER.SHOP
SELLER.SHOP.OPERATOR
```

## Example

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
