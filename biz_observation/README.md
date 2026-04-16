# biz_observation

`biz_observation` provides a standalone Go module for business observation utilities.

This directory is an independent Go module.

## Files

- `log_util.go`: context-aware logging field helpers
- `metrics_util.go`: metric label helpers and duration observation helper
- `trace_util.go`: trace context and tracer abstraction helpers

## Core Types

- `Logger`
- `MetricsRecorder`
- `Tracer`
- `Span`

## Development

Run tests from the module directory:

```bash
cd biz_observation && go test ./...
```
