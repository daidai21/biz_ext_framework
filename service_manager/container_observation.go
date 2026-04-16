package service_manager

import (
	"context"
	"sync"
	"time"

	"github.com/daidai21/biz_ext_framework/biz_observation"
)

// ObservationContainer manages business observation dependencies.
type ObservationContainer struct {
	mu      sync.RWMutex
	logger  biz_observation.Logger
	metrics biz_observation.MetricsRecorder
	tracer  biz_observation.Tracer
}

func NewObservationContainer() *ObservationContainer {
	return &ObservationContainer{}
}

func (c *ObservationContainer) SetLogger(logger biz_observation.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

func (c *ObservationContainer) SetMetricsRecorder(recorder biz_observation.MetricsRecorder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = recorder
}

func (c *ObservationContainer) SetTracer(tracer biz_observation.Tracer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tracer = tracer
}

func (c *ObservationContainer) Logger() biz_observation.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

func (c *ObservationContainer) MetricsRecorder() biz_observation.MetricsRecorder {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

func (c *ObservationContainer) Tracer() biz_observation.Tracer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tracer
}

func (c *ObservationContainer) Log(ctx context.Context, level biz_observation.LogLevel, msg string, fields ...biz_observation.LogField) {
	biz_observation.Log(ctx, c.Logger(), level, msg, fields...)
}

func (c *ObservationContainer) ObserveDuration(ctx context.Context, name string, start time.Time, labels ...biz_observation.MetricLabel) {
	biz_observation.ObserveDuration(ctx, c.MetricsRecorder(), name, start, labels...)
}

func (c *ObservationContainer) StartSpan(ctx context.Context, name string, attrs ...biz_observation.TraceAttribute) (context.Context, biz_observation.Span) {
	return biz_observation.StartSpan(ctx, c.Tracer(), name, attrs...)
}
