# biz_observation

`biz_observation` 提供了一个用于业务观测能力的独立 Go 模块。

该目录本身是一个独立的 Go module。

## 文件说明

- `log_util.go`：带上下文字段的日志工具
- `metrics_util.go`：指标 label 工具和耗时观测工具
- `trace_util.go`：trace 上下文和 tracer 抽象工具

## 核心类型

- `Logger`
- `MetricsRecorder`
- `Tracer`
- `Span`

## 开发

在模块目录下运行测试：

```bash
cd biz_observation && go test ./...
```
