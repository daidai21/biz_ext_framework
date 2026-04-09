# biz_ext_framework

`biz_ext_framework` is a repository for reusable business extension components.

Go modules are managed inside subdirectories. The current module lives in [`ext_model`](./ext_model).

## Packages

### `ext_model`

`ext_model` provides a generic, concurrency-safe model map abstraction:

- `ExtObj`: value contract with `Key() string`
- `ExtModel[V]`: map behavior interface
- `ExtMap[V]`: default implementation
- `CopyExtMap`: copy helper with `WithDeepCopy` and `WithKeyFilter`

Documentation:

- English: [`ext_model/README.md`](./ext_model/README.md)
- 中文: [`ext_model/README-ZH.md`](./ext_model/README-ZH.md)

### `biz_identity`

`biz_identity` provides a business identity component for e-commerce scenarios:

- `BizIdentity`
- `Parser`
- `Validator`

Documentation:

- English: [`biz_identity/README.md`](./biz_identity/README.md)
- 中文: [`biz_identity/README-ZH.md`](./biz_identity/README-ZH.md)

## Quick Start

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

## Development

Run tests from the module directory:

```bash
cd ext_model && go test ./...
```
