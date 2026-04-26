# tools

`tools` contains repository-level command-line utilities for process graph generation and parsing.

## Intro

Current tools:

- `gen_process_graph`: generate BPMN / DAG / FSM graphs from structured JSON specs
- `parse_process_graph`: parse runtime metric logs and render BPMN / DAG / FSM graphs with aggregated metrics

Both tools support Mermaid and Graphviz DOT output.

## Install

Install from GitHub directly:

```bash
go install github.com/daidai21/biz_ext_framework/tools/gen_process_graph@latest
go install github.com/daidai21/biz_ext_framework/tools/parse_process_graph@latest
```

For local development inside this repository, you can also install from the repository root:

```bash
go install ./tools/gen_process_graph
go install ./tools/parse_process_graph
```

After installation, the binaries can be used directly if your `GOBIN` or `GOPATH/bin` is in `PATH`:

```bash
gen_process_graph -h
parse_process_graph -h
```

## Usage

Generate a graph from a static process spec:

```bash
gen_process_graph -type bpmn -input process.json
gen_process_graph -type dag -input dag.json -format dot -output dag.dot
gen_process_graph -type fsm -input fsm.json
```

Parse runtime logs and render a graph with metrics:

```bash
parse_process_graph -type bpmn -input metrics.jsonl
parse_process_graph -type dag -input dag_metrics.jsonl -format dot -output dag.dot
parse_process_graph -type fsm -input fsm_metrics.jsonl -metrics qps,biz_identity,success_rate,p99
```

Tool-specific docs:

- [`gen_process_graph/README.md`](./gen_process_graph/README.md)
- [`parse_process_graph/README.md`](./parse_process_graph/README.md)

## Example

Example BPMN spec:

```json
{
  "name": "order-flow",
  "layers": [
    {"name": "prepare", "nodes": ["prepare"]},
    {"name": "fanout", "nodes": ["audit", "notify"]},
    {"name": "finalize", "nodes": ["finalize"]}
  ]
}
```

Generate Mermaid output:

```bash
gen_process_graph -type bpmn -input process.json
```

Example runtime FSM metrics:

```json
{"type":"fsm","process":"order-fsm","from":"CREATED","event":"PAY","to":"PAID","qps":50,"success_rate":0.98,"p99_ms":42}
{"type":"fsm","process":"order-fsm","from":"PAID","event":"SHIP","to":"SHIPPED","qps":45,"avg_ms":9}
```

Generate an FSM graph with metrics:

```bash
parse_process_graph -type fsm -input fsm_metrics.jsonl -metrics qps,success_rate,p99,avg
```
