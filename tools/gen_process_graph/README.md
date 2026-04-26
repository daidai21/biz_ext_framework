# gen_process_graph

`gen_process_graph` is a small CLI for rendering BPMN, DAG, and FSM process specs into Mermaid or Graphviz DOT.

## Usage

Install:

```bash
go install github.com/daidai21/biz_ext_framework/tools/gen_process_graph@latest
```

```bash
go run ./tools/gen_process_graph -type bpmn -input process.json
go run ./tools/gen_process_graph -type dag -input dag.json -format dot -output dag.dot
go run ./tools/gen_process_graph -type fsm -input fsm.json
```

Flags:

- `-type bpmn|dag|fsm`
- `-input <file>`
- `-format mermaid|dot` (default: `mermaid`)
- `-output <file>` (default: stdout)
- `-direction LR|TD` for BPMN/DAG Mermaid output (default: `LR`)

## Input Specs

### BPMN

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

### DAG

```json
{
  "name": "order-dag",
  "nodes": [
    {"name": "prepare"},
    {"name": "audit", "depends_on": ["prepare"]},
    {"name": "notify", "depends_on": ["prepare"]},
    {"name": "finalize", "depends_on": ["audit", "notify"]}
  ]
}
```

The DAG input also accepts a raw JSON array of nodes.

### FSM

```json
{
  "name": "order-fsm",
  "initial": "CREATED",
  "transitions": [
    {"from": "CREATED", "event": "PAY", "to": "PAID"},
    {"from": "PAID", "event": "SHIP", "to": "SHIPPED"}
  ]
}
```
