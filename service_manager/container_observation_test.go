package service_manager

import (
	"context"
	"testing"
	"time"

	"github.com/daidai21/biz_ext_framework/biz_observation"
)

type observationTestLogger struct {
	msg string
}

func (l *observationTestLogger) Log(ctx context.Context, level biz_observation.LogLevel, msg string, fields ...biz_observation.LogField) {
	l.msg = msg
}

type observationTestMetrics struct {
	count int
}

func (r *observationTestMetrics) Count(ctx context.Context, name string, value int64, labels ...biz_observation.MetricLabel) {
}

func (r *observationTestMetrics) Gauge(ctx context.Context, name string, value float64, labels ...biz_observation.MetricLabel) {
}

func (r *observationTestMetrics) Histogram(ctx context.Context, name string, value float64, labels ...biz_observation.MetricLabel) {
	r.count++
}

type observationTestSpan struct {
	ended bool
}

func (s *observationTestSpan) SetAttributes(attrs ...biz_observation.TraceAttribute) {}

func (s *observationTestSpan) RecordError(err error) {}

func (s *observationTestSpan) End() {
	s.ended = true
}

type observationTestTracer struct {
	span *observationTestSpan
}

func (t *observationTestTracer) StartSpan(ctx context.Context, name string, attrs ...biz_observation.TraceAttribute) (context.Context, biz_observation.Span) {
	t.span = &observationTestSpan{}
	return ctx, t.span
}

func TestObservationContainer(t *testing.T) {
	container := NewObservationContainer()

	logger := &observationTestLogger{}
	metrics := &observationTestMetrics{}
	tracer := &observationTestTracer{}

	container.SetLogger(logger)
	container.SetMetricsRecorder(metrics)
	container.SetTracer(tracer)

	container.Log(context.Background(), biz_observation.LogLevelInfo, "created")
	container.ObserveDuration(context.Background(), "latency", time.Now().Add(-time.Second))
	_, span := container.StartSpan(context.Background(), "trace")
	span.End()

	if logger.msg != "created" {
		t.Fatalf("unexpected logger msg: %s", logger.msg)
	}
	if metrics.count != 1 {
		t.Fatalf("expected 1 histogram, got %d", metrics.count)
	}
	if tracer.span == nil || !tracer.span.ended {
		t.Fatalf("expected ended span")
	}
}
