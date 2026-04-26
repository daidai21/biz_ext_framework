# biz_component

`biz_component` 提供了一个用于 IOC 风格业务组件管理的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 核心概念

- `Container`：IOC 容器
- `GlobalScope`：全局单例对象
- `SessionScope`：按 `biz_ctx.BizSessionId` 隔离的 Session 级对象
- `Key[T]`：带类型的组件 key
- `Provider[T]`：支持依赖解析的泛型延迟构造函数
- `Resolver`：Provider 内部用于查找依赖对象的接口
- `Namespace`：组件所属分层命名空间

## 行为说明

- 全局对象在同一个容器内只会构建一次
- Session 级对象会按 session id 分别构建
- 全局单例统一使用 `GlobalScope` / `GlobalKey` / `RegisterGlobal` / `GlobalObject`
- 同一个组件名可以同时注册为 `GlobalScope` 和 `SessionScope`
- `Resolve(ctx, resolver, key)` 会按 `key.Scope()` 精确解析，直接调用 `ResolveAny` 时若上下文里有 session 则优先取 Session 版本
- Provider 内部通过泛型 helper 解析依赖组件
- 支持循环依赖检测
- 支持分层 namespace 依赖约束
- 容器是并发安全的

## Namespace 分层

`biz_component` 提供以下 namespace：

- `infra`
- `repository`
- `service`
- `domain`
- `capability`
- `business`
- `handler`

依赖规则如下：

- `infra` 不能依赖其他
- `repository` 不能依赖其他
- `service` 仅可依赖 `infra`、`repository`
- `domain` 仅可依赖 `service`、`infra`、`repository`
- `capability` / `business` 仅可依赖 `domain`、`service`、`repository`、`infra`
- `handler` 可以依赖所有

如果不显式指定 namespace，`GlobalKey` / `SessionKey` 默认使用 `handler`。

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
	configKey := biz_component.GlobalKeyIn[string](biz_component.ServiceNamespace, "config")
	componentKey := biz_component.SessionKeyIn[string](biz_component.HandlerNamespace, "order_component")

	_ = biz_component.RegisterGlobal(container, configKey, func(ctx context.Context, resolver biz_component.Resolver) (string, error) {
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
