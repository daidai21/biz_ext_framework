package biz_process

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type testExtension struct {
	beforeCount int
	afterCount  int
	errorCount  int
	trace       []string
}

func (e *testExtension) BeforeTransition(ctx context.Context, from State, to State, event Event, payload any) error {
	e.beforeCount++
	e.trace = append(e.trace, "before")
	return nil
}

func (e *testExtension) AfterTransition(ctx context.Context, from State, to State, event Event, payload any) {
	e.afterCount++
	e.trace = append(e.trace, "after")
}

func (e *testExtension) OnTransitionError(ctx context.Context, from State, to State, event Event, payload any, err error) {
	e.errorCount++
	e.trace = append(e.trace, "error")
}

func mustFSM(t *testing.T, initial State, transitions []Transition, exts ...Extension) *FSM {
	t.Helper()
	fsm, err := NewFSM(initial, transitions, exts...)
	if err != nil {
		t.Fatalf("new fsm failed: %v", err)
	}
	return fsm
}

func TestFSMFireSuccess(t *testing.T) {
	ext := &testExtension{}
	fsm := mustFSM(t, "CREATED", []Transition{
		{From: "CREATED", Event: "PAY", To: "PAID"},
	}, ext)

	state, err := fsm.Fire(context.Background(), "PAY", nil)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if state != "PAID" || fsm.State() != "PAID" {
		t.Fatalf("expected PAID, got state=%q current=%q", state, fsm.State())
	}
	if ext.beforeCount != 1 || ext.afterCount != 1 || ext.errorCount != 0 {
		t.Fatalf("unexpected extension count: before=%d after=%d error=%d", ext.beforeCount, ext.afterCount, ext.errorCount)
	}
}

func TestFSMTransitionNotFound(t *testing.T) {
	ext := &testExtension{}
	fsm := mustFSM(t, "CREATED", []Transition{}, ext)

	state, err := fsm.Fire(context.Background(), "PAY", nil)
	if err == nil {
		t.Fatal("expected transition not found error")
	}
	if !errors.Is(err, ErrTransitionNotFound) {
		t.Fatalf("expected ErrTransitionNotFound, got %v", err)
	}
	if state != "CREATED" || fsm.State() != "CREATED" {
		t.Fatalf("state should keep CREATED, got state=%q current=%q", state, fsm.State())
	}
	if ext.errorCount != 1 {
		t.Fatalf("expected error hook once, got %d", ext.errorCount)
	}
}

func TestFSMGuardRejected(t *testing.T) {
	fsm := mustFSM(t, "CREATED", []Transition{
		{
			From:  "CREATED",
			Event: "PAY",
			To:    "PAID",
			Guard: func(ctx context.Context, from State, event Event, payload any) error {
				return errors.New("balance not enough")
			},
		},
	})

	state, err := fsm.Fire(context.Background(), "PAY", nil)
	if err == nil {
		t.Fatal("expected guard reject error")
	}
	if !errors.Is(err, ErrGuardRejected) {
		t.Fatalf("expected ErrGuardRejected, got %v", err)
	}
	if !strings.Contains(err.Error(), "balance not enough") {
		t.Fatalf("expected guard message in err, got %v", err)
	}
	if state != "CREATED" || fsm.State() != "CREATED" {
		t.Fatalf("state should keep CREATED, got state=%q current=%q", state, fsm.State())
	}
}

func TestFSMActionFailed(t *testing.T) {
	fsm := mustFSM(t, "CREATED", []Transition{
		{
			From:  "CREATED",
			Event: "PAY",
			To:    "PAID",
			Action: func(ctx context.Context, from State, to State, event Event, payload any) error {
				return errors.New("db timeout")
			},
		},
	})

	state, err := fsm.Fire(context.Background(), "PAY", nil)
	if err == nil {
		t.Fatal("expected action failed")
	}
	if !strings.Contains(err.Error(), "action failed") {
		t.Fatalf("expected action failed prefix, got %v", err)
	}
	if state != "CREATED" || fsm.State() != "CREATED" {
		t.Fatalf("state should keep CREATED, got state=%q current=%q", state, fsm.State())
	}
}

func TestFSMExtensionOrder(t *testing.T) {
	ext := &testExtension{}
	fsm := mustFSM(t, "CREATED", []Transition{{From: "CREATED", Event: "PAY", To: "PAID"}}, ext)

	_, err := fsm.Fire(context.Background(), "PAY", nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ext.trace) != 2 || ext.trace[0] != "before" || ext.trace[1] != "after" {
		t.Fatalf("unexpected trace: %#v", ext.trace)
	}
}

func TestNewFSMDuplicateTransition(t *testing.T) {
	_, err := NewFSM("CREATED", []Transition{
		{From: "CREATED", Event: "PAY", To: "PAID"},
		{From: "CREATED", Event: "PAY", To: "CANCELLED"},
	})
	if err == nil {
		t.Fatal("expected duplicate transition error")
	}
}
