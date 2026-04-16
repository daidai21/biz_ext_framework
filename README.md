# biz_ext_framework

`biz_ext_framework` is a repository for reusable business extension components.

Components are organized by top-level directories. Some directories are already independent Go modules, and some are placeholders reserved for follow-up work.

## Architecture

Modules in this repository can be used in two ways:

- use low-level modules independently, without pulling in other modules
- use `service_manager` as the service-side integration layer that wires several modules together

`service_manager` currently integrates:

- `biz_identity`
- `biz_process`
- `ext_model`

Those lower-level modules do not depend on each other and can still be used separately.

```text
                          +-------------------+
                          |  service_manager  |
                          |   integration     |
                          +-------------------+
                            /       |       \
                           /        |        \
                          v         v         v
                +---------------+ +---------------+ +---------------+
                | biz_identity  | |  biz_process  | |   ext_model   |
                | identity wl   | | multi-process | | model filter  |
                +---------------+ +---------------+ +---------------+

Independent usage:

  biz_identity      biz_process      ext_model      ext_spi      ext_process
       |                 |               |             |             |
       +-----------------+---------------+-------------+-------------+
                         each module can be used alone
```

## Directory Layout

- `biz_ctx/`: placeholder directory for business context components
- `biz_identity/`: independent Go module for business identity abstractions
- `biz_process/`: independent Go module for business process FSM
- `ext_interceptor/`: placeholder directory for extension interceptor components
- `ext_model/`: independent Go module for extension model abstractions
- `ext_process/`: independent Go module for extension process template
- `ext_spi/`: independent Go module for SPI template abstractions
- `service_manager/`: independent Go module for service-side integration and container management
- `Makefile`: repository-level helper targets
- `go.mod`: repository-level Go module definition

## Implemented Modules

### `service_manager`

`service_manager` provides a service-side integration layer built on top of other reusable modules:

- `ServiceManager`: service instance lifecycle management
- `ServiceManagerBuilder`: container initialization and service construction
- `IdentityContainer`: business identity whitelist management
- `ProcessContainer`: multiple named process orchestration management
- `SPIContainer`: extension definition to implementation management
- `ModelContainer`: outbound RPC ext model whitelist filtering

Documentation:

- English: [`service_manager/README.md`](./service_manager/README.md)
- 中文: [`service_manager/README-ZH.md`](./service_manager/README-ZH.md)

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

`biz_process` provides process orchestration components:

- FSM
- BPMN-like serial-layer / parallel-node orchestration
- DAG orchestration

Documentation:

- English: [`biz_process/README.md`](./biz_process/README.md)
- 中文: [`biz_process/README-ZH.md`](./biz_process/README-ZH.md)

### `ext_process`

`ext_process` provides a generic extension process template:

- `Mode` (`Serial`, `Parallel`)
- `Template`
- `MatchFunc`
- `ProcessFunc` (with `continueNext` support in serial mode)

Documentation:

- English: [`ext_process/README.md`](./ext_process/README.md)
- 中文: [`ext_process/README-ZH.md`](./ext_process/README-ZH.md)

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
