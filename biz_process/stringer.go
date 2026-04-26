package biz_process

import (
	"encoding/json"
	"sort"
)

type ProcessStringer interface {
	String() string
}

type DAG []GraphNode

type processJSON struct {
	Type   string             `json:"type"`
	Name   string             `json:"name,omitempty"`
	Layers []processLayerJSON `json:"layers"`
}

type processLayerJSON struct {
	Name  string            `json:"name,omitempty"`
	Nodes []processNodeJSON `json:"nodes"`
}

type processNodeJSON struct {
	Name string `json:"name"`
}

type dagJSON struct {
	Type  string        `json:"type"`
	Name  string        `json:"name,omitempty"`
	Nodes []dagNodeJSON `json:"nodes"`
}

type dagNodeJSON struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type fsmJSON struct {
	Type        string              `json:"type"`
	Initial     State               `json:"initial"`
	Current     State               `json:"current"`
	Transitions []fsmTransitionJSON `json:"transitions"`
}

type fsmTransitionJSON struct {
	From      State `json:"from"`
	Event     Event `json:"event"`
	To        State `json:"to"`
	HasGuard  bool  `json:"has_guard,omitempty"`
	HasAction bool  `json:"has_action,omitempty"`
}

func (p Process) String() string {
	layers := make([]processLayerJSON, 0, len(p.Layers))
	for _, layer := range p.Layers {
		nodes := make([]processNodeJSON, 0, len(layer.Nodes))
		for _, node := range layer.Nodes {
			if node == nil {
				nodes = append(nodes, processNodeJSON{})
				continue
			}
			nodes = append(nodes, processNodeJSON{Name: node.NodeName()})
		}
		layers = append(layers, processLayerJSON{
			Name:  layer.Name,
			Nodes: nodes,
		})
	}
	return mustJSONString(processJSON{
		Type:   "bpmn",
		Name:   p.Name,
		Layers: layers,
	})
}

func (d DAG) String() string {
	nodes := make([]dagNodeJSON, 0, len(d))
	for _, node := range d {
		dependsOn := append([]string(nil), node.DependsOn...)
		sort.Strings(dependsOn)
		nodes = append(nodes, dagNodeJSON{
			Name:      node.Name,
			DependsOn: dependsOn,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	return mustJSONString(dagJSON{
		Type:  "dag",
		Nodes: nodes,
	})
}

func mustJSONString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
