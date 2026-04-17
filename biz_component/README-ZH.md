# biz_component

`biz_component` 提供了一个用于 IOC 风格业务组件管理的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心概念

- `Container`：IOC 容器
- `ServiceScope`：服务实例级单例对象
- `SessionScope`：按 `biz_ctx.BizSessionId` 隔离的 Session 级对象
- `Provider`：支持依赖解析的延迟构造函数
- `Resolver`：Provider 内部用于查找依赖对象的接口

## 行为说明

- 服务级对象在同一个容器内只会构建一次
- Session 级对象会按 session id 分别构建
- Provider 内部可以通过 `Resolver` 解析其他组件
- 支持循环依赖检测
- 容器是并发安全的

## 示例

```go
package main

import (
	"context"
	"fmt"

	"github.com/daidai21/biz_ext_framework/biz_component"
	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func main() {
	container := biz_component.NewContainer()

	_ = container.RegisterService("config", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		return "cfg", nil
	})
	_ = container.RegisterSession("order_component", func(ctx context.Context, resolver biz_component.Resolver) (any, error) {
		cfg, err := resolver.Resolve(ctx, "config")
		if err != nil {
			return nil, err
		}
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return fmt.Sprintf("%s:%s", cfg.(string), session.BizSessionId()), nil
	})

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	value, err := container.Resolve(ctx, "order_component")
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
}
```

## 开发

在模块目录下运行测试：

```bash
cd biz_component && go test ./...
```
