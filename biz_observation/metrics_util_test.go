package biz_observation

import (
	"context"
	"testing"
	"time"
)

type histogramRecord struct {
	name   string
	value  float64
	labels []MetricLabel
}

type testMetricsRecorder struct {
	histograms []histogramRecord
}

func (r *testMetricsRecorder) Count(ctx context.Context, name string, value int64, labels ...MetricLabel) {
}

func (r *testMetricsRecorder) Gauge(ctx context.Context, name string, value float64, labels ...MetricLabel) {
}

func (r *testMetricsRecorder) Histogram(ctx context.Context, name string, value float64, labels ...MetricLabel) {
	r.histograms = append(r.histograms, histogramRecord{
		name:   name,
		value:  value,
		labels: append([]MetricLabel(nil), labels...),
	})
}

func TestMergeMetricLabels(t *testing.T) {
	labels := MergeMetricLabels(
		[]MetricLabel{{Key: "region", Value: "cn"}, {Key: "scene", Value: "order"}},
		[]MetricLabel{{Key: "region", Value: "sg"}},
	)

	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
	if labels[0].Key != "region" || labels[0].Value != "sg" {
		t.Fatalf("unexpected region label: %+v", labels[0])
	}
}

func TestNormalizeMetricName(t *testing.T) {
	got := NormalizeMetricName("order.create", "latency-ms", " p99 ")
	if got != "order_create_latency_ms_p99" {
		t.Fatalf("unexpected metric name: %s", got)
	}
}

func TestObserveDuration(t *testing.T) {
	recorder := &testMetricsRecorder{}

	ObserveDuration(context.Background(), recorder, "order_latency", time.Now().Add(-time.Second), MetricLabel{Key: "scene", Value: "create"})

	if len(recorder.histograms) != 1 {
		t.Fatalf("expected 1 histogram record, got %d", len(recorder.histograms))
	}
	if recorder.histograms[0].name != "order_latency" {
		t.Fatalf("unexpected histogram name: %+v", recorder.histograms[0])
	}
}
