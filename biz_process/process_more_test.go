package biz_process

import (
	"context"
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
