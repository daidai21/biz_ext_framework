package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

type graphType string
type graphFormat string

const (
	typeBPMN graphType = "bpmn"
	typeDAG  graphType = "dag"
	typeFSM  graphType = "fsm"

	formatMermaid graphFormat = "mermaid"
	formatDot     graphFormat = "dot"
)

var nonIdentPattern = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

type namedNode struct {
	Name string `json:"name"`
}

func (n *namedNode) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return errors.New("empty node")
	}
	if data[0] == '"' {
		return json.Unmarshal(data, &n.Name)
	}
	var aux struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	n.Name = aux.Name
	return nil
}

type bpmnSpec struct {
	Name   string          `json:"name"`
	Layers []bpmnLayerSpec `json:"layers"`
}

type bpmnLayerSpec struct {
	Name  string      `json:"name"`
	Nodes []namedNode `json:"nodes"`
}

type dagSpec struct {
	Name  string        `json:"name"`
	Nodes []dagNodeSpec `json:"nodes"`
}

type dagNodeSpec struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"depends_on"`
}

type fsmSpec struct {
	Name        string              `json:"name"`
	Initial     string              `json:"initial"`
	Transitions []fsmTransitionSpec `json:"transitions"`
}

type fsmTransitionSpec struct {
	From  string `json:"from"`
	Event string `json:"event"`
	To    string `json:"to"`
}

type options struct {
	graphType graphType
	format    graphFormat
	input     string
	output    string
	direction string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	opts, err := parseFlags(args)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(opts.input)
	if err != nil {
		return fmt.Errorf("read input failed: %w", err)
	}

	graph, err := generateGraph(data, opts)
	if err != nil {
		return err
	}

	if opts.output == "" {
		_, err = io.WriteString(stdout, graph)
		return err
	}
	if err := os.WriteFile(opts.output, []byte(graph), 0o644); err != nil {
		return fmt.Errorf("write output failed: %w", err)
	}
	_, _ = fmt.Fprintf(stderr, "graph written to %s\n", opts.output)
	return nil
}

func parseFlags(args []string) (options, error) {
	fs := flag.NewFlagSet("gen_process_graph", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts options
	var typ string
	var format string
	fs.StringVar(&typ, "type", "", "graph type: bpmn|dag|fsm")
	fs.StringVar(&format, "format", string(formatMermaid), "output format: mermaid|dot")
	fs.StringVar(&opts.input, "input", "", "input json path")
	fs.StringVar(&opts.output, "output", "", "output path, stdout if empty")
	fs.StringVar(&opts.direction, "direction", "LR", "graph direction for BPMN/DAG mermaid: LR|TD")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	opts.graphType = graphType(strings.ToLower(strings.TrimSpace(typ)))
	opts.format = graphFormat(strings.ToLower(strings.TrimSpace(format)))

	switch opts.graphType {
	case typeBPMN, typeDAG, typeFSM:
	default:
		return opts, fmt.Errorf("invalid -type %q, want bpmn|dag|fsm", typ)
	}
	switch opts.format {
	case formatMermaid, formatDot:
	default:
		return opts, fmt.Errorf("invalid -format %q, want mermaid|dot", format)
	}
	if opts.input == "" {
		return opts, errors.New("-input is required")
	}
	opts.direction = strings.ToUpper(strings.TrimSpace(opts.direction))
	if opts.direction == "" {
		opts.direction = "LR"
	}
	return opts, nil
}

func generateGraph(data []byte, opts options) (string, error) {
	switch opts.graphType {
	case typeBPMN:
		var spec bpmnSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			return "", fmt.Errorf("decode bpmn spec failed: %w", err)
		}
		if err := validateBPMNSpec(spec); err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderBPMNDot(spec), nil
		}
		return renderBPMNMermaid(spec, opts.direction), nil
	case typeDAG:
		spec, err := decodeDAGSpec(data)
		if err != nil {
			return "", err
		}
		if err := validateDAGSpec(spec); err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderDAGDot(spec), nil
		}
		return renderDAGMermaid(spec, opts.direction), nil
	case typeFSM:
		var spec fsmSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			return "", fmt.Errorf("decode fsm spec failed: %w", err)
		}
		if err := validateFSMSpec(spec); err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderFSMDot(spec), nil
		}
		return renderFSMMermaid(spec), nil
	default:
		return "", fmt.Errorf("unsupported type %q", opts.graphType)
	}
}

func decodeDAGSpec(data []byte) (dagSpec, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return dagSpec{}, errors.New("empty dag spec")
	}
	if data[0] == '[' {
		var nodes []dagNodeSpec
		if err := json.Unmarshal(data, &nodes); err != nil {
			return dagSpec{}, fmt.Errorf("decode dag spec failed: %w", err)
		}
		return dagSpec{Nodes: nodes}, nil
	}
	var spec dagSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return dagSpec{}, fmt.Errorf("decode dag spec failed: %w", err)
	}
	return spec, nil
}

func validateBPMNSpec(spec bpmnSpec) error {
	if len(spec.Layers) == 0 {
		return errors.New("bpmn spec requires at least one layer")
	}
	seen := map[string]struct{}{}
	for i, layer := range spec.Layers {
		if len(layer.Nodes) == 0 {
			return fmt.Errorf("bpmn layer[%d] must contain at least one node", i)
		}
		for j, node := range layer.Nodes {
			if strings.TrimSpace(node.Name) == "" {
				return fmt.Errorf("bpmn layer[%d] node[%d] name is required", i, j)
			}
			if _, ok := seen[node.Name]; ok {
				return fmt.Errorf("bpmn duplicate node name %q", node.Name)
			}
			seen[node.Name] = struct{}{}
		}
	}
	return nil
}

func validateDAGSpec(spec dagSpec) error {
	if len(spec.Nodes) == 0 {
		return errors.New("dag spec requires at least one node")
	}
	seen := map[string]struct{}{}
	for i, node := range spec.Nodes {
		if strings.TrimSpace(node.Name) == "" {
			return fmt.Errorf("dag node[%d] name is required", i)
		}
		if _, ok := seen[node.Name]; ok {
			return fmt.Errorf("dag duplicate node name %q", node.Name)
		}
		seen[node.Name] = struct{}{}
	}
	for _, node := range spec.Nodes {
		for _, dep := range node.DependsOn {
			if _, ok := seen[dep]; !ok {
				return fmt.Errorf("dag node %q depends on unknown node %q", node.Name, dep)
			}
		}
	}
	return nil
}

func validateFSMSpec(spec fsmSpec) error {
	if len(spec.Transitions) == 0 {
		return errors.New("fsm spec requires at least one transition")
	}
	for i, t := range spec.Transitions {
		if strings.TrimSpace(t.From) == "" || strings.TrimSpace(t.Event) == "" || strings.TrimSpace(t.To) == "" {
			return fmt.Errorf("fsm transition[%d] requires from/event/to", i)
		}
	}
	return nil
}

func renderBPMNMermaid(spec bpmnSpec, direction string) string {
	var b strings.Builder
	b.WriteString("flowchart ")
	b.WriteString(direction)
	b.WriteByte('\n')
	if spec.Name != "" {
		fmt.Fprintf(&b, "    %% process: %s\n", spec.Name)
	}

	nodeIDs := map[string]string{}
	for i, layer := range spec.Layers {
		layerID := fmt.Sprintf("layer_%d", i)
		fmt.Fprintf(&b, "    subgraph %s[%q]\n", layerID, chooseLabel(layer.Name, fmt.Sprintf("layer-%d", i)))
		for j, node := range layer.Nodes {
			id := mermaidID(node.Name, i, j)
			nodeIDs[node.Name] = id
			fmt.Fprintf(&b, "        %s[%q]\n", id, node.Name)
		}
		b.WriteString("    end\n")
	}
	for i := 0; i < len(spec.Layers)-1; i++ {
		for _, from := range spec.Layers[i].Nodes {
			for _, to := range spec.Layers[i+1].Nodes {
				fmt.Fprintf(&b, "    %s --> %s\n", nodeIDs[from.Name], nodeIDs[to.Name])
			}
		}
	}
	return b.String()
}

func renderBPMNDot(spec bpmnSpec) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if spec.Name != "" {
		fmt.Fprintf(&b, "    label=%q;\n", spec.Name)
	}
	nodeIDs := map[string]string{}
	for i, layer := range spec.Layers {
		fmt.Fprintf(&b, "    subgraph cluster_%d {\n", i)
		fmt.Fprintf(&b, "        label=%q;\n", chooseLabel(layer.Name, fmt.Sprintf("layer-%d", i)))
		for j, node := range layer.Nodes {
			id := dotID(node.Name, i, j)
			nodeIDs[node.Name] = id
			fmt.Fprintf(&b, "        %s [label=%q];\n", id, node.Name)
		}
		b.WriteString("    }\n")
	}
	for i := 0; i < len(spec.Layers)-1; i++ {
		for _, from := range spec.Layers[i].Nodes {
			for _, to := range spec.Layers[i+1].Nodes {
				fmt.Fprintf(&b, "    %s -> %s;\n", nodeIDs[from.Name], nodeIDs[to.Name])
			}
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderDAGMermaid(spec dagSpec, direction string) string {
	var b strings.Builder
	b.WriteString("flowchart ")
	b.WriteString(direction)
	b.WriteByte('\n')
	if spec.Name != "" {
		fmt.Fprintf(&b, "    %% dag: %s\n", spec.Name)
	}
	nodeIDs := map[string]string{}
	for i, node := range spec.Nodes {
		id := mermaidID(node.Name, i, 0)
		nodeIDs[node.Name] = id
		fmt.Fprintf(&b, "    %s[%q]\n", id, node.Name)
	}

	roots := make([]string, 0)
	for _, node := range spec.Nodes {
		if len(node.DependsOn) == 0 {
			roots = append(roots, node.Name)
			continue
		}
		for _, dep := range node.DependsOn {
			fmt.Fprintf(&b, "    %s --> %s\n", nodeIDs[dep], nodeIDs[node.Name])
		}
	}
	sort.Strings(roots)
	for _, root := range roots {
		fmt.Fprintf(&b, "    %% root: %s\n", root)
	}
	return b.String()
}

func renderDAGDot(spec dagSpec) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if spec.Name != "" {
		fmt.Fprintf(&b, "    label=%q;\n", spec.Name)
	}
	nodeIDs := map[string]string{}
	for i, node := range spec.Nodes {
		id := dotID(node.Name, i, 0)
		nodeIDs[node.Name] = id
		fmt.Fprintf(&b, "    %s [label=%q];\n", id, node.Name)
	}
	for _, node := range spec.Nodes {
		for _, dep := range node.DependsOn {
			fmt.Fprintf(&b, "    %s -> %s;\n", nodeIDs[dep], nodeIDs[node.Name])
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderFSMMermaid(spec fsmSpec) string {
	var b strings.Builder
	b.WriteString("stateDiagram-v2\n")
	if spec.Name != "" {
		fmt.Fprintf(&b, "    %% fsm: %s\n", spec.Name)
	}
	if spec.Initial != "" {
		fmt.Fprintf(&b, "    [*] --> %s\n", mermaidStateID(spec.Initial))
	}
	seen := map[string]bool{}
	for _, t := range spec.Transitions {
		from := mermaidStateID(t.From)
		to := mermaidStateID(t.To)
		if !seen[t.From] {
			fmt.Fprintf(&b, "    state %s as %q\n", from, t.From)
			seen[t.From] = true
		}
		if !seen[t.To] {
			fmt.Fprintf(&b, "    state %s as %q\n", to, t.To)
			seen[t.To] = true
		}
		fmt.Fprintf(&b, "    %s --> %s: %s\n", from, to, t.Event)
	}
	return b.String()
}

func renderFSMDot(spec fsmSpec) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if spec.Name != "" {
		fmt.Fprintf(&b, "    label=%q;\n", spec.Name)
	}
	if spec.Initial != "" {
		b.WriteString("    __start__ [shape=point];\n")
		fmt.Fprintf(&b, "    __start__ -> %s;\n", dotStateID(spec.Initial))
	}
	seen := map[string]bool{}
	for _, t := range spec.Transitions {
		from := dotStateID(t.From)
		to := dotStateID(t.To)
		if !seen[t.From] {
			fmt.Fprintf(&b, "    %s [label=%q];\n", from, t.From)
			seen[t.From] = true
		}
		if !seen[t.To] {
			fmt.Fprintf(&b, "    %s [label=%q];\n", to, t.To)
			seen[t.To] = true
		}
		fmt.Fprintf(&b, "    %s -> %s [label=%q];\n", from, to, t.Event)
	}
	b.WriteString("}\n")
	return b.String()
}

func chooseLabel(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func mermaidID(name string, indexes ...int) string {
	base := sanitizeIdent(name)
	if base == "" {
		base = "node"
	}
	for _, idx := range indexes {
		base += fmt.Sprintf("_%d", idx)
	}
	return base
}

func dotID(name string, indexes ...int) string {
	return mermaidID(name, indexes...)
}

func mermaidStateID(name string) string {
	base := sanitizeIdent(name)
	if base == "" {
		base = "state"
	}
	return "state_" + base
}

func dotStateID(name string) string {
	return mermaidStateID(name)
}

func sanitizeIdent(name string) string {
	sanitized := nonIdentPattern.ReplaceAllString(strings.TrimSpace(name), "_")
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		return ""
	}
	if sanitized[0] >= '0' && sanitized[0] <= '9' {
		return "n_" + sanitized
	}
	return sanitized
}
