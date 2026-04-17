package biz_observation

import (
	"context"
	"testing"
	"time"
)

func TestLogHelpersEdgeCases(t *testing.T) {
	ctx := WithLogFields(context.Background())
	if fields := LogFieldsFromContext(nil); fields != nil {
		t.Fatalf("expected nil fields from nil ctx, got %#v", fields)
	}

	fields := MergeLogFields(
		[]LogField{{Key: "", Value: "skip"}, {Key: "a", Value: 1}},
		[]LogField{{Key: "a", Value: 2}, {Key: "b", Value: 3}},
	)
	if len(fields) != 2 || fields[0].Value != 2 || fields[1].Key != "b" {
		t.Fatalf("unexpected merged log fields: %#v", fields)
	}

	Log(ctx, nil, LogLevelInfo, "noop")
}

func TestMetricsHelpersEdgeCases(t *testing.T) {
	labels := MergeMetricLabels(
		[]MetricLabel{{Key: "", Value: "skip"}, {Key: "region", Value: "cn"}},
		[]MetricLabel{{Key: "region", Value: "sg"}},
	)
	if len(labels) != 1 || labels[0].Value != "sg" {
		t.Fatalf("unexpected merged labels: %#v", labels)
	}

	if got := NormalizeMetricName(" order.create ", "", "latency-ms", " p99 "); got != "order_create_latency_ms_p99" {
		t.Fatalf("unexpected normalized metric name: %s", got)
	}

	recorder := &testMetricsRecorder{}
	ObserveDuration(context.Background(), nil, "metric", time.Now())
	ObserveDuration(context.Background(), recorder, "", time.Now())
	ObserveDuration(context.Background(), recorder, "metric", time.Time{})
	if len(recorder.histograms) != 0 {
		t.Fatalf("expected no histogram records for ignored calls, got %d", len(recorder.histograms))
	}
}

func TestTraceHelpersEdgeCases(t *testing.T) {
	if info, ok := TraceInfoFromContext(nil); ok || info != (TraceInfo{}) {
		t.Fatalf("expected empty trace info from nil ctx, got %+v %v", info, ok)
	}
	if traceID := CurrentTraceID(context.Background()); traceID != "" {
		t.Fatalf("expected empty trace id, got %q", traceID)
	}

	var span noopSpan
	span.SetAttributes(TraceAttribute{Key: "k", Value: "v"})
	span.RecordError(nil)
	span.End()
}
