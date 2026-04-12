# biz_ext_framework

`biz_ext_framework` is a repository for reusable business extension components.

Components are organized by top-level directories. Some directories are already independent Go modules, and some are placeholders reserved for follow-up work.

## Directory Layout

- `biz_ctx/`: placeholder directory for business context components
- `biz_identity/`: independent Go module for business identity abstractions
- `biz_process/`: independent Go module for business process FSM
- `ext_interceptor/`: placeholder directory for extension interceptor components
- `ext_model/`: independent Go module for extension model abstractions
- `ext_process/`: placeholder directory for extension process components
- `ext_spi/`: independent Go module for SPI template abstractions
- `service_manager/`: placeholder directory for service manager components
- `Makefile`: repository-level helper targets
- `go.mod`: repository-level Go module definition

## Implemented Modules

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

`biz_identity` provides a technical component for business identity abstractions:

- `BizIdentity`
- `Parser`
- `Validator`

Documentation:

- English: [`biz_identity/README.md`](./biz_identity/README.md)
- 中文: [`biz_identity/README-ZH.md`](./biz_identity/README-ZH.md)

### `biz_process`

`biz_process` provides an extensible FSM framework:

- `State` / `Event`
- `Transition` (`From + Event -> To`)
- `Guard`
- `Action`
- `Extension` hooks

Documentation:

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

### `ext_spi`

`ext_spi` provides a generic SPI template with four modes:

- `First`
- `All`
- `FirstMatched`
- `AllMatched`

Documentation:

- English: [`ext_spi/README.md`](./ext_spi/README.md)
- 中文: [`ext_spi/README-ZH.md`](./ext_spi/README-ZH.md)

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

Run tests from the target module directory:

```bash
cd ext_model && go test ./...
```

Repository-level helper target:

```bash
make statistics_lines
```
