package biz_observation

import (
	"context"
	"strings"
	"time"
)

type MetricLabel struct {
	Key   string
	Value string
}

type MetricsRecorder interface {
	Count(ctx context.Context, name string, value int64, labels ...MetricLabel)
	Gauge(ctx context.Context, name string, value float64, labels ...MetricLabel)
	Histogram(ctx context.Context, name string, value float64, labels ...MetricLabel)
}

func MergeMetricLabels(groups ...[]MetricLabel) []MetricLabel {
	type labelIndex struct {
		label MetricLabel
	}

	ordered := make([]labelIndex, 0)
	indexByKey := map[string]int{}
	for _, group := range groups {
		for _, label := range group {
			if label.Key == "" {
				continue
			}
			if idx, ok := indexByKey[label.Key]; ok {
				ordered[idx].label = label
				continue
			}
			indexByKey[label.Key] = len(ordered)
			ordered = append(ordered, labelIndex{label: label})
		}
	}

	labels := make([]MetricLabel, 0, len(ordered))
	for _, item := range ordered {
		labels = append(labels, item.label)
	}
	return labels
}

func NormalizeMetricName(parts ...string) string {
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		part = strings.ReplaceAll(part, "-", "_")
		part = strings.ReplaceAll(part, ".", "_")
		part = strings.ReplaceAll(part, " ", "_")
		names = append(names, part)
	}
	return strings.Join(names, "_")
}

func ObserveDuration(ctx context.Context, recorder MetricsRecorder, name string, start time.Time, labels ...MetricLabel) {
	if recorder == nil || name == "" || start.IsZero() {
		return
	}
	recorder.Histogram(ctx, name, time.Since(start).Seconds(), labels...)
}
