package biz_ctx

import (
	"context"
	"testing"
)

type testSession struct {
	id string
}

func (s testSession) BizSessionId() string {
	return s.id
}

type testInstance struct {
	id   string
	name string
}

func (i testInstance) BizInstanceId() string {
	return i.id
}

var _ BizSession = testSession{}
var _ BizInstance = testInstance{}
var _ BizCtx = (*Ctx)(nil)

func TestCtxOneSessionToManyInstances(t *testing.T) {
	ctx := NewBizCtx()
	baseCtx := WithBizSession(context.Background(), testSession{id: "s1"})

	if !ctx.Set(baseCtx, testInstance{id: "i1", name: "order"}) {
		t.Fatal("expected set i1 success")
	}
	if !ctx.Set(baseCtx, testInstance{id: "i2", name: "refund"}) {
		t.Fatal("expected set i2 success")
	}

	v1, ok := ctx.Get(baseCtx, "i1")
	if !ok {
		t.Fatal("expected i1 to exist under session s1")
	}
	i1, ok := v1.(testInstance)
	if !ok || i1.name != "order" {
		t.Fatalf("unexpected i1: %#v", v1)
	}

	v2, ok := ctx.Get(baseCtx, "i2")
	if !ok {
		t.Fatal("expected i2 to exist under session s1")
	}
	i2, ok := v2.(testInstance)
	if !ok || i2.name != "refund" {
		t.Fatalf("unexpected i2: %#v", v2)
	}
}

func TestCtxSessionIsolation(t *testing.T) {
	ctx := NewBizCtx()

	s1ctx := WithBizSession(context.Background(), testSession{id: "s1"})
	s2ctx := WithBizSession(context.Background(), testSession{id: "s2"})
	ctx.Set(s1ctx, testInstance{id: "i1", name: "order"})
	ctx.Set(s2ctx, testInstance{id: "i1", name: "payment"})

	v1, ok := ctx.Get(s1ctx, "i1")
	if !ok {
		t.Fatal("expected s1/i1")
	}
	i1 := v1.(testInstance)
	if i1.name != "order" {
		t.Fatalf("expected order, got %s", i1.name)
	}

	v2, ok := ctx.Get(s2ctx, "i1")
	if !ok {
		t.Fatal("expected s2/i1")
	}
	i2 := v2.(testInstance)
	if i2.name != "payment" {
		t.Fatalf("expected payment, got %s", i2.name)
	}
}

func TestCtxDel(t *testing.T) {
	ctx := NewBizCtx()
	baseCtx := WithBizSession(context.Background(), testSession{id: "s1"})
	ctx.Set(baseCtx, testInstance{id: "i1", name: "order"})

	deleted, ok := ctx.Del(baseCtx, "i1")
	if !ok {
		t.Fatal("expected delete success")
	}
	instance := deleted.(testInstance)
	if instance.name != "order" {
		t.Fatalf("expected order, got %s", instance.name)
	}

	_, ok = ctx.Get(baseCtx, "i1")
	if ok {
		t.Fatal("expected i1 removed")
	}
}

func TestCtxForEachAndList(t *testing.T) {
	ctx := NewBizCtx()
	baseCtx := WithBizSession(context.Background(), testSession{id: "s1"})
	ctx.Set(baseCtx, testInstance{id: "i1", name: "order"})
	ctx.Set(baseCtx, testInstance{id: "i2", name: "refund"})

	list := ctx.List(baseCtx)
	if len(list) != 2 {
		t.Fatalf("expected list size 2, got %d", len(list))
	}

	seen := map[string]bool{}
	ctx.ForEach(baseCtx, func(instance BizInstance) {
		seen[instance.BizInstanceId()] = true
	})
	if len(seen) != 2 || !seen["i1"] || !seen["i2"] {
		t.Fatalf("unexpected foreach set: %#v", seen)
	}
}

func TestCtxSetWithoutBizSessionInContext(t *testing.T) {
	ctx := NewBizCtx()
	ok := ctx.Set(context.Background(), testInstance{id: "i1", name: "order"})
	if ok {
		t.Fatal("expected set fail when session missing in context")
	}
}

func TestCtxMethodsWithoutBizSessionInContext(t *testing.T) {
	ctx := NewBizCtx()
	noSessionCtx := context.Background()

	if _, ok := ctx.Get(noSessionCtx, "i1"); ok {
		t.Fatal("expected get fail when session missing in context")
	}

	if _, ok := ctx.Del(noSessionCtx, "i1"); ok {
		t.Fatal("expected del fail when session missing in context")
	}

	if list := ctx.List(noSessionCtx); list != nil {
		t.Fatalf("expected nil list, got %#v", list)
	}

	count := 0
	ctx.ForEach(noSessionCtx, func(instance BizInstance) {
		count++
	})
	if count != 0 {
		t.Fatalf("expected foreach no-op, got %d", count)
	}
}

func TestBizSessionContextHelpers(t *testing.T) {
	_, ok := BizSessionFromContext(context.Background())
	if ok {
		t.Fatal("expected no session in empty context")
	}

	withSession := WithBizSession(context.Background(), testSession{id: "s1"})
	session, ok := BizSessionFromContext(withSession)
	if !ok {
		t.Fatal("expected session in context")
	}
	if session.BizSessionId() != "s1" {
		t.Fatalf("expected session id s1, got %s", session.BizSessionId())
	}
}
