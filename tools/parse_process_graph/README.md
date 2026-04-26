# parse_process_graph

`parse_process_graph` parses runtime metric logs and renders BPMN, DAG, or FSM dependency graphs in Mermaid or Graphviz DOT.

## Usage

```bash
go run ./tools/parse_process_graph -type bpmn -input metrics.jsonl
go run ./tools/parse_process_graph -type dag -input dag_metrics.jsonl -format dot -output dag.dot
go run ./tools/parse_process_graph -type fsm -input fsm_metrics.jsonl -metrics qps,biz_identity,success_rate,p99
```

Flags:

- `-type bpmn|dag|fsm`
- `-input <file>`: JSONL or JSON array
- `-process <name>`: optional process filter when one file contains multiple processes
- `-format mermaid|dot` (default: `mermaid`)
- `-output <file>` (default: stdout)
- `-direction LR|TD` for BPMN/DAG Mermaid output
- `-metrics qps,biz_identity,success_rate,avg,p90,p99`

## Supported Input Fields

Common fields:

- `type`: `bpmn` / `dag` / `fsm`
- `process` or `process_name`
- `biz_identity`
- `qps`
- `success_rate` or `sr`
- `avg_ms` or `avg`
- `p90_ms` or `p90`
- `p99_ms` or `p99`
- `latency_ms`: `{ "avg": 10, "p90": 20, "p99": 30 }`

### BPMN

```json
{"type":"bpmn","process":"order-flow","layer":"prepare","node":"prepare","qps":120,"success_rate":0.99,"avg_ms":10,"p90_ms":18,"p99_ms":30,"biz_identity":"SELLER.SHOP"}
{"type":"bpmn","process":"order-flow","layer":"fanout","node":"audit","qps":80,"success_rate":0.995,"avg_ms":12,"biz_identity":"SELLER.BIZ"}
```

Required topology fields:

- `layer`
- `node` or `node_name`

### DAG

```json
{"type":"dag","process":"order-dag","node":"prepare","qps":100}
{"type":"dag","process":"order-dag","node":"audit","depends_on":["prepare"],"qps":80,"avg_ms":15}
{"type":"dag","process":"order-dag","node":"notify","depends_on":["prepare"],"qps":75,"avg_ms":11}
```

Required topology fields:

- `node` or `node_name`
- optional `depends_on`

### FSM

```json
{"type":"fsm","process":"order-fsm","from":"CREATED","event":"PAY","to":"PAID","qps":60,"success_rate":0.98,"p99_ms":42}
{"type":"fsm","process":"order-fsm","from":"PAID","event":"SHIP","to":"SHIPPED","qps":55,"success_rate":0.995,"avg_ms":9}
```

Required topology fields:

- `from`
- `event`
- `to`

## Aggregation

When multiple log lines hit the same BPMN node, DAG node, or FSM transition, the tool aggregates them:

- `qps`: sum
- `biz_identity`: unique sorted set
- `success_rate`, `avg`, `p90`, `p99`: weighted average by `qps` when available, otherwise arithmetic average
