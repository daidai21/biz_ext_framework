# biz_component

`biz_component` 提供了一个用于 IOC 风格业务组件管理的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心概念

- `Container`：IOC 容器
- `ServiceScope`：服务实例级单例对象
- `SessionScope`：按 `biz_ctx.BizSessionId` 隔离的 Session 级对象
- `Key[T]`：带类型的组件 key
- `Provider[T]`：支持依赖解析的泛型延迟构造函数
- `Resolver`：Provider 内部用于查找依赖对象的接口

## 行为说明

- 服务级对象在同一个容器内只会构建一次
- Session 级对象会按 session id 分别构建
- Provider 内部通过泛型 helper 解析依赖组件
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
	configKey := biz_component.ServiceKey[string]("config")
	componentKey := biz_component.SessionKey[string]("order_component")

	_ = biz_component.RegisterService(container, configKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		return "cfg", nil
	})
	_ = biz_component.RegisterSession(container, componentKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
		cfg, err := biz_component.Resolve(ctx, resolver, configKey)
		if err != nil {
			return "", err
		}
		session, _ := biz_ctx.BizSessionFromContext(ctx)
		return fmt.Sprintf("%s:%s", cfg, session.BizSessionId()), nil
	})

	ctx := biz_ctx.WithBizSession(context.Background(), biz_ctx.NewBizSession("s1"))
	value, err := biz_component.Resolve(ctx, container, componentKey)
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
