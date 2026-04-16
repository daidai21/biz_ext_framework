package service_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

func TestCtxContainer(t *testing.T) {
	container := NewCtxContainer()

	session, err := container.Create("s1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if session.BizSessionId() != "s1" {
		t.Fatalf("unexpected session id: %s", session.BizSessionId())
	}

	ctx, err := container.WithSession(context.Background(), "s1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	fromCtx, ok := container.SessionFromContext(ctx)
	if !ok || fromCtx.BizSessionId() != "s1" {
		t.Fatalf("expected session from ctx")
	}
}

func TestCtxContainerRegisterInvalid(t *testing.T) {
	container := NewCtxContainer()
	if err := container.Register(nil); !errors.Is(err, ErrNilBizSession) {
		t.Fatalf("expected ErrNilBizSession, got %v", err)
	}

	if _, err := container.Create(""); !errors.Is(err, ErrInvalidBizSessionID) {
		t.Fatalf("expected ErrInvalidBizSessionID, got %v", err)
	}

	if err := container.Register(biz_ctx.NewBizSession("")); !errors.Is(err, ErrInvalidBizSessionID) {
		t.Fatalf("expected ErrInvalidBizSessionID, got %v", err)
	}
}
