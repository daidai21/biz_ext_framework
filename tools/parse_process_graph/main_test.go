package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildBPMNGraphFromJSONL(t *testing.T) {
	events := []metricEvent{
		{
			Type:        typeBPMN,
			Process:     "order-flow",
			Node:        "prepare",
			Layer:       "prepare",
			BizIdentity: "SELLER.SHOP",
			QPS:         floatPtr(120),
			SuccessRate: floatPtr(0.99),
			AvgMS:       floatPtr(10),
			P90MS:       floatPtr(18),
			P99MS:       floatPtr(30),
		},
		{
			Type:        typeBPMN,
			Process:     "order-flow",
			Node:        "audit",
			Layer:       "fanout",
			BizIdentity: "SELLER.BIZ",
			QPS:         floatPtr(80),
			SuccessRate: floatPtr(99),
			AvgMS:       floatPtr(12),
		},
	}

	graph, err := buildGraph(events, options{
		graphType: typeBPMN,
		format:    formatMermaid,
		direction: "LR",
		metrics:   []metricName{metricQPS, metricBizIdentity, metricSuccessRate, metricAvg, metricP90, metricP99},
	})
	if err != nil {
		t.Fatalf("build graph failed: %v", err)
	}

	for _, want := range []string{
		"flowchart LR",
		"prepare_0_0",
		"QPS: 120.00",
		"biz_identity: SELLER.SHOP",
		"SR: 99.00%",
		"audit_1_0",
	} {
		if !strings.Contains(graph, want) {
			t.Fatalf("expected graph to contain %q, got:\n%s", want, graph)
		}
	}
}

func TestBuildDAGGraphProcessFilter(t *testing.T) {
	events := []metricEvent{
		{Type: typeDAG, Process: "order", Node: "prepare"},
		{Type: typeDAG, Process: "order", Node: "audit", DependsOn: []string{"prepare"}, QPS: floatPtr(20)},
		{Type: typeDAG, Process: "refund", Node: "prepare"},
	}

	graph, err := buildGraph(events, options{
		graphType: typeDAG,
		format:    formatDot,
		process:   "order",
		metrics:   []metricName{metricQPS},
	})
	if err != nil {
		t.Fatalf("build graph failed: %v", err)
	}
	if strings.Contains(graph, "refund") {
		t.Fatalf("expected refund process filtered out, got:\n%s", graph)
	}
	if !strings.Contains(graph, "audit_0_0") || !strings.Contains(graph, "QPS: 20.00") {
		t.Fatalf("expected audit node metrics in dot output, got:\n%s", graph)
	}
}

func TestBuildFSMGraphWithMetrics(t *testing.T) {
	events := []metricEvent{
		{Type: typeFSM, Process: "order-fsm", From: "CREATED", Event: "PAY", To: "PAID", QPS: floatPtr(50), SuccessRate: floatPtr(0.98), P99MS: floatPtr(42)},
	}

	graph, err := buildGraph(events, options{
		graphType: typeFSM,
		format:    formatMermaid,
		metrics:   []metricName{metricQPS, metricSuccessRate, metricP99},
	})
	if err != nil {
		t.Fatalf("build graph failed: %v", err)
	}
	for _, want := range []string{
		"stateDiagram-v2",
		"state_CREATED --> state_PAID: PAY\\nQPS: 50.00\\nSR: 98.00%\\nP99: 42.00ms",
	} {
		if !strings.Contains(graph, want) {
			t.Fatalf("expected graph to contain %q, got:\n%s", want, graph)
		}
	}
}

func TestRunParseProcessGraph(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(input, []byte("{\"type\":\"dag\",\"process\":\"order\",\"node\":\"prepare\"}\n{\"type\":\"dag\",\"process\":\"order\",\"node\":\"audit\",\"depends_on\":[\"prepare\"],\"qps\":10}\n"), 0o644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	err := run([]string{"-type", "dag", "-input", input, "-metrics", "qps"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "prepare") || !strings.Contains(stdout.String(), "audit") {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestParseMetricsValidation(t *testing.T) {
	if _, err := parseMetrics("qps,bad"); err == nil {
		t.Fatal("expected invalid metrics error")
	}
}
