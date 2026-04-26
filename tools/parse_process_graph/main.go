package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
)

type graphType string
type graphFormat string
type metricName string

const (
	typeBPMN graphType = "bpmn"
	typeDAG  graphType = "dag"
	typeFSM  graphType = "fsm"

	formatMermaid graphFormat = "mermaid"
	formatDot     graphFormat = "dot"

	metricQPS         metricName = "qps"
	metricBizIdentity metricName = "biz_identity"
	metricSuccessRate metricName = "success_rate"
	metricAvg         metricName = "avg"
	metricP90         metricName = "p90"
	metricP99         metricName = "p99"
)

var nonIdentPattern = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

type options struct {
	graphType graphType
	format    graphFormat
	input     string
	output    string
	process   string
	direction string
	metrics   []metricName
}

type metricEvent struct {
	Type        graphType
	Process     string
	Node        string
	Layer       string
	DependsOn   []string
	From        string
	Event       string
	To          string
	BizIdentity string
	QPS         *float64
	SuccessRate *float64
	AvgMS       *float64
	P90MS       *float64
	P99MS       *float64
}

type metricAggregate struct {
	qpsSum             float64
	successWeightedSum float64
	avgWeightedSum     float64
	p90WeightedSum     float64
	p99WeightedSum     float64
	weightSum          float64
	sampleCount        int
	identities         map[string]struct{}
}

type bpmnNode struct {
	Name    string
	Layer   string
	Metrics metricAggregate
}

type dagNode struct {
	Name      string
	DependsOn []string
	Metrics   metricAggregate
}

type fsmEdge struct {
	From    string
	Event   string
	To      string
	Metrics metricAggregate
}

type parsedBPMN struct {
	Process string
	Layers  []string
	Nodes   map[string]bpmnNode
}

type parsedDAG struct {
	Process string
	Nodes   map[string]dagNode
}

type parsedFSM struct {
	Process string
	Edges   []fsmEdge
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

	events, err := loadEvents(opts.input)
	if err != nil {
		return err
	}

	graph, err := buildGraph(events, opts)
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
	fs := flag.NewFlagSet("parse_process_graph", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts options
	var typ string
	var format string
	var metrics string
	fs.StringVar(&typ, "type", "", "graph type: bpmn|dag|fsm")
	fs.StringVar(&format, "format", string(formatMermaid), "output format: mermaid|dot")
	fs.StringVar(&opts.input, "input", "", "input log path, supports jsonl/json array")
	fs.StringVar(&opts.output, "output", "", "output path, stdout if empty")
	fs.StringVar(&opts.process, "process", "", "optional process name filter")
	fs.StringVar(&opts.direction, "direction", "LR", "graph direction for BPMN/DAG mermaid: LR|TD")
	fs.StringVar(&metrics, "metrics", "qps,success_rate,avg,p90,p99", "comma-separated metrics: qps,biz_identity,success_rate,avg,p90,p99")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	opts.graphType = graphType(strings.ToLower(strings.TrimSpace(typ)))
	opts.format = graphFormat(strings.ToLower(strings.TrimSpace(format)))
	opts.direction = strings.ToUpper(strings.TrimSpace(opts.direction))
	if opts.direction == "" {
		opts.direction = "LR"
	}

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
	parsedMetrics, err := parseMetrics(metrics)
	if err != nil {
		return opts, err
	}
	opts.metrics = parsedMetrics
	return opts, nil
}

func parseMetrics(raw string) ([]metricName, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	items := strings.Split(raw, ",")
	seen := map[metricName]struct{}{}
	metrics := make([]metricName, 0, len(items))
	for _, item := range items {
		name := metricName(strings.ToLower(strings.TrimSpace(item)))
		if name == "latency" {
			name = metricAvg
		}
		switch name {
		case metricQPS, metricBizIdentity, metricSuccessRate, metricAvg, metricP90, metricP99:
		default:
			return nil, fmt.Errorf("unsupported metric %q", item)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		metrics = append(metrics, name)
	}
	return metrics, nil
}

func loadEvents(path string) ([]metricEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read input failed: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty input")
	}
	if data[0] == '[' {
		var rawItems []json.RawMessage
		if err := json.Unmarshal(data, &rawItems); err != nil {
			return nil, fmt.Errorf("decode json array failed: %w", err)
		}
		events := make([]metricEvent, 0, len(rawItems))
		for i, raw := range rawItems {
			event, err := decodeMetricEvent(raw)
			if err != nil {
				return nil, fmt.Errorf("decode event[%d] failed: %w", i, err)
			}
			events = append(events, event)
		}
		return events, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	events := make([]metricEvent, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		event, err := decodeMetricEvent(line)
		if err != nil {
			return nil, fmt.Errorf("decode line %d failed: %w", lineNo, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan input failed: %w", err)
	}
	if len(events) == 0 {
		return nil, errors.New("no events found in input")
	}
	return events, nil
}

func decodeMetricEvent(data []byte) (metricEvent, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return metricEvent{}, err
	}

	event := metricEvent{
		Type:        graphType(strings.ToLower(readString(raw, "type"))),
		Process:     firstNonEmpty(readString(raw, "process"), readString(raw, "process_name"), readString(raw, "name")),
		Node:        firstNonEmpty(readString(raw, "node"), readString(raw, "node_name"), readString(raw, "step"), readString(raw, "name")),
		Layer:       firstNonEmpty(readString(raw, "layer"), readString(raw, "layer_name")),
		From:        firstNonEmpty(readString(raw, "from"), readString(raw, "from_state")),
		Event:       firstNonEmpty(readString(raw, "event"), readString(raw, "trigger")),
		To:          firstNonEmpty(readString(raw, "to"), readString(raw, "to_state")),
		BizIdentity: firstNonEmpty(readString(raw, "biz_identity"), readString(raw, "identity")),
		DependsOn:   readStringSlice(raw, "depends_on", "deps"),
		QPS:         readFloatPointer(raw, "qps"),
		SuccessRate: readFloatPointer(raw, "success_rate", "sr"),
		AvgMS:       readFloatPointer(raw, "avg_ms", "avg"),
		P90MS:       readFloatPointer(raw, "p90_ms", "p90"),
		P99MS:       readFloatPointer(raw, "p99_ms", "p99"),
	}

	if latencyRaw, ok := raw["latency_ms"]; ok {
		var latency map[string]float64
		if err := json.Unmarshal(latencyRaw, &latency); err == nil {
			if event.AvgMS == nil {
				if v, ok := latency["avg"]; ok {
					event.AvgMS = floatPtr(v)
				}
			}
			if event.P90MS == nil {
				if v, ok := latency["p90"]; ok {
					event.P90MS = floatPtr(v)
				}
			}
			if event.P99MS == nil {
				if v, ok := latency["p99"]; ok {
					event.P99MS = floatPtr(v)
				}
			}
		}
	}

	return event, nil
}

func buildGraph(events []metricEvent, opts options) (string, error) {
	filtered := make([]metricEvent, 0, len(events))
	processNames := map[string]struct{}{}
	for _, event := range events {
		if event.Type != opts.graphType {
			continue
		}
		if opts.process != "" && event.Process != opts.process {
			continue
		}
		filtered = append(filtered, event)
		if event.Process != "" {
			processNames[event.Process] = struct{}{}
		}
	}
	if len(filtered) == 0 {
		return "", errors.New("no matching events found")
	}
	if opts.process == "" && len(processNames) > 1 {
		names := make([]string, 0, len(processNames))
		for name := range processNames {
			names = append(names, name)
		}
		sort.Strings(names)
		return "", fmt.Errorf("multiple processes found %v, please specify -process", names)
	}

	switch opts.graphType {
	case typeBPMN:
		parsed, err := parseBPMN(filtered)
		if err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderBPMNDot(parsed, opts.metrics), nil
		}
		return renderBPMNMermaid(parsed, opts.metrics, opts.direction), nil
	case typeDAG:
		parsed, err := parseDAG(filtered)
		if err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderDAGDot(parsed, opts.metrics), nil
		}
		return renderDAGMermaid(parsed, opts.metrics, opts.direction), nil
	case typeFSM:
		parsed, err := parseFSM(filtered)
		if err != nil {
			return "", err
		}
		if opts.format == formatDot {
			return renderFSMDot(parsed, opts.metrics), nil
		}
		return renderFSMMermaid(parsed, opts.metrics), nil
	default:
		return "", fmt.Errorf("unsupported graph type %q", opts.graphType)
	}
}

func parseBPMN(events []metricEvent) (parsedBPMN, error) {
	result := parsedBPMN{Nodes: map[string]bpmnNode{}}
	layerIndex := map[string]int{}
	for _, event := range events {
		if event.Node == "" {
			return parsedBPMN{}, errors.New("bpmn event requires node or name")
		}
		if event.Layer == "" {
			return parsedBPMN{}, fmt.Errorf("bpmn event node %q requires layer", event.Node)
		}
		if result.Process == "" {
			result.Process = event.Process
		}
		node := result.Nodes[event.Node]
		node.Name = event.Node
		node.Layer = event.Layer
		node.Metrics.add(event)
		result.Nodes[event.Node] = node
		if _, ok := layerIndex[event.Layer]; !ok {
			layerIndex[event.Layer] = len(layerIndex)
			result.Layers = append(result.Layers, event.Layer)
		}
	}
	return result, nil
}

func parseDAG(events []metricEvent) (parsedDAG, error) {
	result := parsedDAG{Nodes: map[string]dagNode{}}
	for _, event := range events {
		if event.Node == "" {
			return parsedDAG{}, errors.New("dag event requires node or name")
		}
		if result.Process == "" {
			result.Process = event.Process
		}
		node := result.Nodes[event.Node]
		node.Name = event.Node
		if len(event.DependsOn) > 0 {
			node.DependsOn = append([]string(nil), event.DependsOn...)
			sort.Strings(node.DependsOn)
		}
		node.Metrics.add(event)
		result.Nodes[event.Node] = node
	}
	return result, nil
}

func parseFSM(events []metricEvent) (parsedFSM, error) {
	result := parsedFSM{Edges: make([]fsmEdge, 0)}
	index := map[string]int{}
	for _, event := range events {
		if event.From == "" || event.Event == "" || event.To == "" {
			return parsedFSM{}, errors.New("fsm event requires from, event, and to")
		}
		if result.Process == "" {
			result.Process = event.Process
		}
		key := event.From + "\x00" + event.Event + "\x00" + event.To
		pos, ok := index[key]
		if !ok {
			pos = len(result.Edges)
			index[key] = pos
			result.Edges = append(result.Edges, fsmEdge{
				From:  event.From,
				Event: event.Event,
				To:    event.To,
			})
		}
		result.Edges[pos].Metrics.add(event)
	}
	sort.Slice(result.Edges, func(i, j int) bool {
		if result.Edges[i].From != result.Edges[j].From {
			return result.Edges[i].From < result.Edges[j].From
		}
		if result.Edges[i].Event != result.Edges[j].Event {
			return result.Edges[i].Event < result.Edges[j].Event
		}
		return result.Edges[i].To < result.Edges[j].To
	})
	return result, nil
}

func (m *metricAggregate) add(event metricEvent) {
	weight := 1.0
	if event.QPS != nil && *event.QPS > 0 {
		weight = *event.QPS
		m.qpsSum += *event.QPS
	}
	if event.SuccessRate != nil {
		m.successWeightedSum += normalizePercent(*event.SuccessRate) * weight
	}
	if event.AvgMS != nil {
		m.avgWeightedSum += *event.AvgMS * weight
	}
	if event.P90MS != nil {
		m.p90WeightedSum += *event.P90MS * weight
	}
	if event.P99MS != nil {
		m.p99WeightedSum += *event.P99MS * weight
	}
	m.weightSum += weight
	m.sampleCount++
	if event.BizIdentity != "" {
		if m.identities == nil {
			m.identities = map[string]struct{}{}
		}
		m.identities[event.BizIdentity] = struct{}{}
	}
}

func (m metricAggregate) metricLines(selected []metricName) []string {
	lines := make([]string, 0, len(selected))
	for _, metric := range selected {
		switch metric {
		case metricQPS:
			if m.qpsSum > 0 {
				lines = append(lines, fmt.Sprintf("QPS: %.2f", m.qpsSum))
			}
		case metricBizIdentity:
			if len(m.identities) > 0 {
				lines = append(lines, "biz_identity: "+strings.Join(m.sortedIdentities(), ","))
			}
		case metricSuccessRate:
			if m.weightSum > 0 && m.successWeightedSum > 0 {
				lines = append(lines, fmt.Sprintf("SR: %.2f%%", m.successWeightedSum/m.weightSum*100))
			}
		case metricAvg:
			if m.weightSum > 0 && m.avgWeightedSum > 0 {
				lines = append(lines, fmt.Sprintf("AVG: %.2fms", m.avgWeightedSum/m.weightSum))
			}
		case metricP90:
			if m.weightSum > 0 && m.p90WeightedSum > 0 {
				lines = append(lines, fmt.Sprintf("P90: %.2fms", m.p90WeightedSum/m.weightSum))
			}
		case metricP99:
			if m.weightSum > 0 && m.p99WeightedSum > 0 {
				lines = append(lines, fmt.Sprintf("P99: %.2fms", m.p99WeightedSum/m.weightSum))
			}
		}
	}
	return lines
}

func (m metricAggregate) sortedIdentities() []string {
	ids := make([]string, 0, len(m.identities))
	for id := range m.identities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func renderBPMNMermaid(parsed parsedBPMN, metrics []metricName, direction string) string {
	var b strings.Builder
	b.WriteString("flowchart ")
	b.WriteString(direction)
	b.WriteByte('\n')
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    %% process: %s\n", parsed.Process)
	}
	layers := orderedBPMNLayers(parsed)
	nodeIDs := map[string]string{}
	for i, layerName := range layers {
		fmt.Fprintf(&b, "    subgraph layer_%d[%q]\n", i, layerName)
		names := namesInLayer(parsed, layerName)
		for j, name := range names {
			id := mermaidID(name, i, j)
			nodeIDs[name] = id
			fmt.Fprintf(&b, "        %s[%q]\n", id, joinLabel(name, parsed.Nodes[name].Metrics.metricLines(metrics)))
		}
		b.WriteString("    end\n")
	}
	for i := 0; i < len(layers)-1; i++ {
		for _, from := range namesInLayer(parsed, layers[i]) {
			for _, to := range namesInLayer(parsed, layers[i+1]) {
				fmt.Fprintf(&b, "    %s --> %s\n", nodeIDs[from], nodeIDs[to])
			}
		}
	}
	return b.String()
}

func renderBPMNDot(parsed parsedBPMN, metrics []metricName) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    label=%q;\n", parsed.Process)
	}
	layers := orderedBPMNLayers(parsed)
	nodeIDs := map[string]string{}
	for i, layerName := range layers {
		fmt.Fprintf(&b, "    subgraph cluster_%d {\n", i)
		fmt.Fprintf(&b, "        label=%q;\n", layerName)
		names := namesInLayer(parsed, layerName)
		for j, name := range names {
			id := dotID(name, i, j)
			nodeIDs[name] = id
			fmt.Fprintf(&b, "        %s [label=%q];\n", id, joinLabel(name, parsed.Nodes[name].Metrics.metricLines(metrics)))
		}
		b.WriteString("    }\n")
	}
	for i := 0; i < len(layers)-1; i++ {
		for _, from := range namesInLayer(parsed, layers[i]) {
			for _, to := range namesInLayer(parsed, layers[i+1]) {
				fmt.Fprintf(&b, "    %s -> %s;\n", nodeIDs[from], nodeIDs[to])
			}
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderDAGMermaid(parsed parsedDAG, metrics []metricName, direction string) string {
	var b strings.Builder
	b.WriteString("flowchart ")
	b.WriteString(direction)
	b.WriteByte('\n')
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    %% process: %s\n", parsed.Process)
	}
	nodeIDs := map[string]string{}
	names := orderedDAGNames(parsed)
	for i, name := range names {
		id := mermaidID(name, i, 0)
		nodeIDs[name] = id
		fmt.Fprintf(&b, "    %s[%q]\n", id, joinLabel(name, parsed.Nodes[name].Metrics.metricLines(metrics)))
	}
	for _, name := range names {
		for _, dep := range parsed.Nodes[name].DependsOn {
			fmt.Fprintf(&b, "    %s --> %s\n", nodeIDs[dep], nodeIDs[name])
		}
	}
	return b.String()
}

func renderDAGDot(parsed parsedDAG, metrics []metricName) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    label=%q;\n", parsed.Process)
	}
	nodeIDs := map[string]string{}
	names := orderedDAGNames(parsed)
	for i, name := range names {
		id := dotID(name, i, 0)
		nodeIDs[name] = id
		fmt.Fprintf(&b, "    %s [label=%q];\n", id, joinLabel(name, parsed.Nodes[name].Metrics.metricLines(metrics)))
	}
	for _, name := range names {
		for _, dep := range parsed.Nodes[name].DependsOn {
			fmt.Fprintf(&b, "    %s -> %s;\n", nodeIDs[dep], nodeIDs[name])
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderFSMMermaid(parsed parsedFSM, metrics []metricName) string {
	var b strings.Builder
	b.WriteString("stateDiagram-v2\n")
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    %% process: %s\n", parsed.Process)
	}
	states := map[string]struct{}{}
	for _, edge := range parsed.Edges {
		states[edge.From] = struct{}{}
		states[edge.To] = struct{}{}
	}
	orderedStates := make([]string, 0, len(states))
	for state := range states {
		orderedStates = append(orderedStates, state)
	}
	sort.Strings(orderedStates)
	for _, state := range orderedStates {
		fmt.Fprintf(&b, "    state %s as %q\n", mermaidStateID(state), state)
	}
	for _, edge := range parsed.Edges {
		label := joinLabel(edge.Event, edge.Metrics.metricLines(metrics))
		fmt.Fprintf(&b, "    %s --> %s: %s\n", mermaidStateID(edge.From), mermaidStateID(edge.To), label)
	}
	return b.String()
}

func renderFSMDot(parsed parsedFSM, metrics []metricName) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("    rankdir=LR;\n")
	if parsed.Process != "" {
		fmt.Fprintf(&b, "    label=%q;\n", parsed.Process)
	}
	states := map[string]struct{}{}
	for _, edge := range parsed.Edges {
		states[edge.From] = struct{}{}
		states[edge.To] = struct{}{}
	}
	orderedStates := make([]string, 0, len(states))
	for state := range states {
		orderedStates = append(orderedStates, state)
	}
	sort.Strings(orderedStates)
	for _, state := range orderedStates {
		fmt.Fprintf(&b, "    %s [label=%q];\n", dotStateID(state), state)
	}
	for _, edge := range parsed.Edges {
		fmt.Fprintf(&b, "    %s -> %s [label=%q];\n", dotStateID(edge.From), dotStateID(edge.To), joinLabel(edge.Event, edge.Metrics.metricLines(metrics)))
	}
	b.WriteString("}\n")
	return b.String()
}

func orderedBPMNLayers(parsed parsedBPMN) []string {
	layers := append([]string(nil), parsed.Layers...)
	return layers
}

func namesInLayer(parsed parsedBPMN, layer string) []string {
	names := make([]string, 0)
	for name, node := range parsed.Nodes {
		if node.Layer == layer {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func orderedDAGNames(parsed parsedDAG) []string {
	names := make([]string, 0, len(parsed.Nodes))
	for name := range parsed.Nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func joinLabel(title string, metrics []string) string {
	if len(metrics) == 0 {
		return title
	}
	return title + `\n` + strings.Join(metrics, `\n`)
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

func readString(raw map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(value, &s); err == nil {
			return s
		}
	}
	return ""
}

func readStringSlice(raw map[string]json.RawMessage, keys ...string) []string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var items []string
		if err := json.Unmarshal(value, &items); err == nil {
			return items
		}
	}
	return nil
}

func readFloatPointer(raw map[string]json.RawMessage, keys ...string) *float64 {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var f float64
		if err := json.Unmarshal(value, &f); err == nil {
			return &f
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizePercent(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	if value > 1 {
		return value / 100
	}
	return value
}

func floatPtr(v float64) *float64 {
	return &v
}
