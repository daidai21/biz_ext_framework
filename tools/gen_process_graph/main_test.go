package main

import (
	"strings"
	"testing"
)

func TestGenerateBPMNMermaid(t *testing.T) {
	input := []byte(`{
		"name":"order-flow",
		"layers":[
			{"name":"prepare","nodes":["prepare"]},
			{"name":"fanout","nodes":["audit","notify"]},
			{"name":"finalize","nodes":["finalize"]}
		]
	}`)

	graph, err := generateGraph(input, options{graphType: typeBPMN, format: formatMermaid, direction: "LR"})
	if err != nil {
		t.Fatalf("generate graph failed: %v", err)
	}

	for _, want := range []string{
		"flowchart LR",
		`subgraph layer_0["prepare"]`,
		"prepare_0_0 --> audit_1_0",
		"notify_1_1 --> finalize_2_0",
	} {
		if !strings.Contains(graph, want) {
			t.Fatalf("expected graph to contain %q, got:\n%s", want, graph)
		}
	}
}

func TestGenerateDAGDotFromArray(t *testing.T) {
	input := []byte(`[
		{"name":"prepare"},
		{"name":"audit","depends_on":["prepare"]},
		{"name":"notify","depends_on":["prepare"]}
	]`)

	graph, err := generateGraph(input, options{graphType: typeDAG, format: formatDot, direction: "LR"})
	if err != nil {
		t.Fatalf("generate graph failed: %v", err)
	}

	for _, want := range []string{
		"digraph G {",
		"prepare_0_0 -> audit_1_0;",
		"prepare_0_0 -> notify_2_0;",
	} {
		if !strings.Contains(graph, want) {
			t.Fatalf("expected graph to contain %q, got:\n%s", want, graph)
		}
	}
}

func TestGenerateFSMMermaid(t *testing.T) {
	input := []byte(`{
		"name":"order-fsm",
		"initial":"CREATED",
		"transitions":[
			{"from":"CREATED","event":"PAY","to":"PAID"},
			{"from":"PAID","event":"SHIP","to":"SHIPPED"}
		]
	}`)

	graph, err := generateGraph(input, options{graphType: typeFSM, format: formatMermaid})
	if err != nil {
		t.Fatalf("generate graph failed: %v", err)
	}

	for _, want := range []string{
		"stateDiagram-v2",
		"[*] --> state_CREATED",
		"state_CREATED --> state_PAID: PAY",
		`state state_SHIPPED as "SHIPPED"`,
	} {
		if !strings.Contains(graph, want) {
			t.Fatalf("expected graph to contain %q, got:\n%s", want, graph)
		}
	}
}

func TestParseFlagsValidation(t *testing.T) {
	if _, err := parseFlags([]string{"-type", "bad", "-input", "x.json"}); err == nil {
		t.Fatal("expected invalid type error")
	}
	if _, err := parseFlags([]string{"-type", "dag"}); err == nil {
		t.Fatal("expected missing input error")
	}
}
