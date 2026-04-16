package biz_observation

import "context"

type TraceAttribute struct {
	Key   string
	Value any
}

type TraceInfo struct {
	TraceID string
	SpanID  string
}

type Span interface {
	SetAttributes(attrs ...TraceAttribute)
	RecordError(err error)
	End()
}

type Tracer interface {
	StartSpan(ctx context.Context, name string, attrs ...TraceAttribute) (context.Context, Span)
}

type traceInfoContextKey struct{}

type noopSpan struct{}

func (noopSpan) SetAttributes(attrs ...TraceAttribute) {}

func (noopSpan) RecordError(err error) {}

func (noopSpan) End() {}

func WithTraceInfo(ctx context.Context, info TraceInfo) context.Context {
	return context.WithValue(ctx, traceInfoContextKey{}, info)
}

func TraceInfoFromContext(ctx context.Context) (TraceInfo, bool) {
	if ctx == nil {
		return TraceInfo{}, false
	}
	info, ok := ctx.Value(traceInfoContextKey{}).(TraceInfo)
	return info, ok
}

func CurrentTraceID(ctx context.Context) string {
	info, ok := TraceInfoFromContext(ctx)
	if !ok {
		return ""
	}
	return info.TraceID
}

func StartSpan(ctx context.Context, tracer Tracer, name string, attrs ...TraceAttribute) (context.Context, Span) {
	if tracer == nil {
		return ctx, noopSpan{}
	}
	return tracer.StartSpan(ctx, name, attrs...)
}
