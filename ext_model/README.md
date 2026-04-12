# ext_model

`ext_model` provides a small generic map container for business models.

This directory is an independent Go module.

## Core Types

### `ExtObj`

Any value stored in `ExtMap` must implement:

```go
type ExtObj interface {
    Key() string
}
```

`Key()` is the only source of truth for the map key.

### `ExtModel[V ExtObj]`

`ExtModel` is the behavior interface implemented by `ExtMap`.

```go
type ExtModel[V ExtObj] interface {
    Get(key string) (V, bool)
    Set(value V)
    Del(key string) (V, bool)
    ForEach(fn func(value V))
}
```

### `ExtMap[V ExtObj]`

`ExtMap` is a concurrency-safe container built on `map[string]V`.

Public methods:

- `Set(value V)`
- `Get(key string) (V, bool)`
- `Del(key string) (V, bool)`
- `ForEach(fn func(value V))`

## Behavior

- The zero value of `ExtMap` is ready to use.
- `Set` always writes by `value.Key()`.
- `Get` and `Del` operate by explicit string key.
- `ForEach` iterates over a snapshot of current values.
- Iteration order is not guaranteed.

Because `ForEach` uses a snapshot, calling `Set` or `Del` inside the callback will not break the current iteration.

## Example: Attach Multiple Extension Structs to `UserDO`

The following example shows how to attach multiple extension records to one user object.

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
	info.ForEach(func(value ext_model.ExtObj) {
		t.Log(value)
	})
}
```

## Copy Utility

If you need to copy an `ExtMap`, use `ext_model.CopyExtMap`:

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

Available options:

- `WithDeepCopy`
- `WithKeyFilter`
