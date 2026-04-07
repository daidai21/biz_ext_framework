# biz_ext_framework

`biz_ext_framework` is a Go workspace for reusable business extension components.

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

## Quick Start

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

## Development

Run tests:

```bash
go test ./...
```
