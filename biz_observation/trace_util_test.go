package biz_observation

import (
	"context"
	"errors"
	"testing"
)

type testSpan struct {
	attrs []TraceAttribute
	err   error
	ended bool
}

func (s *testSpan) SetAttributes(attrs ...TraceAttribute) {
	s.attrs = append(s.attrs, attrs...)
}

func (s *testSpan) RecordError(err error) {
	s.err = err
}

func (s *testSpan) End() {
	s.ended = true
}

type testTracer struct {
	ctx  context.Context
	span *testSpan
}

func (t *testTracer) StartSpan(ctx context.Context, name string, attrs ...TraceAttribute) (context.Context, Span) {
	t.ctx = WithTraceInfo(ctx, TraceInfo{TraceID: "trace-1", SpanID: "span-1"})
	t.span = &testSpan{}
	t.span.SetAttributes(attrs...)
	return t.ctx, t.span
}

func TestTraceInfo(t *testing.T) {
	ctx := WithTraceInfo(context.Background(), TraceInfo{TraceID: "trace-1", SpanID: "span-1"})

	info, ok := TraceInfoFromContext(ctx)
	if !ok {
		t.Fatal("expected trace info")
	}
	if info.TraceID != "trace-1" || CurrentTraceID(ctx) != "trace-1" {
		t.Fatalf("unexpected trace info: %+v", info)
	}
}

func TestStartSpan(t *testing.T) {
	tracer := &testTracer{}

	ctx, span := StartSpan(context.Background(), tracer, "create_order", TraceAttribute{Key: "scene", Value: "order"})
	info, ok := TraceInfoFromContext(ctx)
	if !ok || info.TraceID != "trace-1" {
		t.Fatalf("expected trace info in context, got %+v", info)
	}

	span.RecordError(errors.New("boom"))
	span.End()

	testSpan, ok := span.(*testSpan)
	if !ok {
		t.Fatalf("expected testSpan, got %T", span)
	}
	if len(testSpan.attrs) != 1 || testSpan.attrs[0].Key != "scene" {
		t.Fatalf("unexpected attrs: %+v", testSpan.attrs)
	}
	if testSpan.err == nil || !testSpan.ended {
		t.Fatalf("expected span record error and end")
	}
}

func TestStartSpanNoop(t *testing.T) {
	ctx, span := StartSpan(context.Background(), nil, "noop")
	if ctx == nil || span == nil {
		t.Fatal("expected noop span")
	}
	span.End()
}
