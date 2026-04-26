package biz_process

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type beforeFailExtension struct{}

func (beforeFailExtension) BeforeTransition(context.Context, State, State, Event, any) error {
	return errors.New("before failed")
}

func (beforeFailExtension) AfterTransition(context.Context, State, State, Event, any) {}

func (beforeFailExtension) OnTransitionError(context.Context, State, State, Event, any, error) {}

func TestTaskRunRequiresTask(t *testing.T) {
	err := Task{Name: "prepare"}.Run(context.Background())
	if !errors.Is(err, ErrInvalidProcess) {
		t.Fatalf("expected ErrInvalidProcess, got %v", err)
	}
}

func TestActionNodeName(t *testing.T) {
	var action Action
	if action.NodeName() != "action" {
		t.Fatalf("unexpected action node name: %q", action.NodeName())
	}
}

func TestFSMBeforeTransitionError(t *testing.T) {
	fsm, err := NewFSM("CREATED", []Transition{
		{From: "CREATED", Event: "PAY", To: "PAID"},
	}, beforeFailExtension{})
	if err != nil {
		t.Fatalf("new fsm failed: %v", err)
	}

	state, err := fsm.Fire(context.Background(), "PAY", nil)
	if err == nil {
		t.Fatal("expected before transition hook error")
	}
	if state != "CREATED" {
		t.Fatalf("expected state to remain CREATED, got %q", state)
	}
}

func TestNoopExtensionMethods(t *testing.T) {
	var ext NoopExtension
	if err := ext.BeforeTransition(context.Background(), "a", "b", "evt", nil); err != nil {
		t.Fatalf("expected noop before transition to succeed, got %v", err)
	}
	ext.AfterTransition(context.Background(), "a", "b", "evt", nil)
	ext.OnTransitionError(context.Background(), "a", "b", "evt", nil, errors.New("boom"))
}

func TestProcessStringJSON(t *testing.T) {
	process := Process{
		Name: "order-flow",
		Layers: []ProcessLayer{
			{
				Name: "prepare",
				Nodes: []ProcessNode{
					Task{Name: "prepare", Task: func(context.Context) error { return nil }},
				},
			},
			{
				Name: "fanout",
				Nodes: []ProcessNode{
					Task{Name: "audit", Task: func(context.Context) error { return nil }},
					Task{Name: "notify", Task: func(context.Context) error { return nil }},
				},
			},
		},
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(process.String()), &decoded); err != nil {
		t.Fatalf("unmarshal process string failed: %v", err)
	}
	if decoded["type"] != "bpmn" {
		t.Fatalf("unexpected process type: %v", decoded["type"])
	}
	if decoded["name"] != "order-flow" {
		t.Fatalf("unexpected process name: %v", decoded["name"])
	}
}

func TestDAGStringJSON(t *testing.T) {
	dag := DAG{
		{Name: "notify", DependsOn: []string{"prepare", "audit"}},
		{Name: "prepare"},
		{Name: "audit", DependsOn: []string{"prepare"}},
	}

	want := `{"type":"dag","nodes":[{"name":"audit","depends_on":["prepare"]},{"name":"notify","depends_on":["audit","prepare"]},{"name":"prepare"}]}`
	if dag.String() != want {
		t.Fatalf("unexpected dag string:\nwant: %s\ngot:  %s", want, dag.String())
	}
}

func TestFSMStringJSON(t *testing.T) {
	fsm, err := NewFSM("CREATED", []Transition{
		{From: "PAID", Event: "SHIP", To: "SHIPPED", Action: func(context.Context, State, State, Event, any) error { return nil }},
		{From: "CREATED", Event: "PAY", To: "PAID", Guard: func(context.Context, State, Event, any) error { return nil }},
	})
	if err != nil {
		t.Fatalf("new fsm failed: %v", err)
	}
	if _, err := fsm.Fire(context.Background(), "PAY", nil); err != nil {
		t.Fatalf("fsm fire failed: %v", err)
	}

	want := `{"type":"fsm","initial":"CREATED","current":"PAID","transitions":[{"from":"CREATED","event":"PAY","to":"PAID","has_guard":true},{"from":"PAID","event":"SHIP","to":"SHIPPED","has_action":true}]}`
	if fsm.String() != want {
		t.Fatalf("unexpected fsm string:\nwant: %s\ngot:  %s", want, fsm.String())
	}
}
