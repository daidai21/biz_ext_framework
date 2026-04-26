package biz_process

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type State string
type Event string

var (
	ErrTransitionNotFound = errors.New("transition not found")
	ErrGuardRejected      = errors.New("guard rejected")
)

type Guard func(ctx context.Context, from State, event Event, payload any) error

type Action func(ctx context.Context, from State, to State, event Event, payload any) error

func (Action) NodeName() string {
	return "action"
}

type Transition struct {
	From   State
	Event  Event
	To     State
	Guard  Guard
	Action Action
}

type Extension interface {
	BeforeTransition(ctx context.Context, from State, to State, event Event, payload any) error
	AfterTransition(ctx context.Context, from State, to State, event Event, payload any)
	OnTransitionError(ctx context.Context, from State, to State, event Event, payload any, err error)
}

type NoopExtension struct{}

func (NoopExtension) BeforeTransition(context.Context, State, State, Event, any) error   { return nil }
func (NoopExtension) AfterTransition(context.Context, State, State, Event, any)          {}
func (NoopExtension) OnTransitionError(context.Context, State, State, Event, any, error) {}

type transitionKey struct {
	from  State
	event Event
}

type FSM struct {
	mu sync.Mutex

	initial     State
	state       State
	rules       map[transitionKey]Transition
	transitions []Transition
	exts        []Extension
}

func NewFSM(initial State, transitions []Transition, extensions ...Extension) (*FSM, error) {
	rules := make(map[transitionKey]Transition, len(transitions))
	for _, t := range transitions {
		key := transitionKey{from: t.From, event: t.Event}
		if _, exists := rules[key]; exists {
			return nil, fmt.Errorf("duplicate transition: from=%q event=%q", t.From, t.Event)
		}
		rules[key] = t
	}

	exts := make([]Extension, 0, len(extensions))
	for _, ext := range extensions {
		if ext != nil {
			exts = append(exts, ext)
		}
	}

	return &FSM{
		initial:     initial,
		state:       initial,
		rules:       rules,
		transitions: append([]Transition(nil), transitions...),
		exts:        exts,
	}, nil
}

func (f *FSM) State() State {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}

func (f *FSM) Fire(ctx context.Context, event Event, payload any) (State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	from := f.state
	t, ok := f.rules[transitionKey{from: from, event: event}]
	if !ok {
		err := fmt.Errorf("%w: from=%q event=%q", ErrTransitionNotFound, from, event)
		f.onTransitionError(ctx, from, "", event, payload, err)
		return from, err
	}

	to := t.To
	for _, ext := range f.exts {
		if err := ext.BeforeTransition(ctx, from, to, event, payload); err != nil {
			wrapped := fmt.Errorf("before transition hook failed: %w", err)
			f.onTransitionError(ctx, from, to, event, payload, wrapped)
			return from, wrapped
		}
	}

	if t.Guard != nil {
		if err := t.Guard(ctx, from, event, payload); err != nil {
			wrapped := fmt.Errorf("%w: %v", ErrGuardRejected, err)
			f.onTransitionError(ctx, from, to, event, payload, wrapped)
			return from, wrapped
		}
	}

	if t.Action != nil {
		if err := t.Action(ctx, from, to, event, payload); err != nil {
			wrapped := fmt.Errorf("action failed: %w", err)
			f.onTransitionError(ctx, from, to, event, payload, wrapped)
			return from, wrapped
		}
	}

	f.state = to
	for _, ext := range f.exts {
		ext.AfterTransition(ctx, from, to, event, payload)
	}
	return to, nil
}

func (f *FSM) onTransitionError(ctx context.Context, from State, to State, event Event, payload any, err error) {
	for _, ext := range f.exts {
		ext.OnTransitionError(ctx, from, to, event, payload, err)
	}
}

func (f *FSM) String() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	transitions := make([]fsmTransitionJSON, 0, len(f.transitions))
	for _, t := range f.transitions {
		transitions = append(transitions, fsmTransitionJSON{
			From:      t.From,
			Event:     t.Event,
			To:        t.To,
			HasGuard:  t.Guard != nil,
			HasAction: t.Action != nil,
		})
	}
	sort.Slice(transitions, func(i, j int) bool {
		if transitions[i].From != transitions[j].From {
			return transitions[i].From < transitions[j].From
		}
		if transitions[i].Event != transitions[j].Event {
			return transitions[i].Event < transitions[j].Event
		}
		return transitions[i].To < transitions[j].To
	})

	return mustJSONString(fsmJSON{
		Type:        "fsm",
		Initial:     f.initial,
		Current:     f.state,
		Transitions: transitions,
	})
}
