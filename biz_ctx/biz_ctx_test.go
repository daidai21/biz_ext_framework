package biz_ctx

import "testing"

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
	session := testSession{id: "s1"}

	ctx.Set(session, testInstance{id: "i1", name: "order"})
	ctx.Set(session, testInstance{id: "i2", name: "refund"})

	v1, ok := ctx.Get("s1", "i1")
	if !ok {
		t.Fatal("expected i1 to exist under session s1")
	}
	i1, ok := v1.(testInstance)
	if !ok || i1.name != "order" {
		t.Fatalf("unexpected i1: %#v", v1)
	}

	v2, ok := ctx.Get("s1", "i2")
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

	ctx.Set(testSession{id: "s1"}, testInstance{id: "i1", name: "order"})
	ctx.Set(testSession{id: "s2"}, testInstance{id: "i1", name: "payment"})

	v1, ok := ctx.Get("s1", "i1")
	if !ok {
		t.Fatal("expected s1/i1")
	}
	i1 := v1.(testInstance)
	if i1.name != "order" {
		t.Fatalf("expected order, got %s", i1.name)
	}

	v2, ok := ctx.Get("s2", "i1")
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
	ctx.Set(testSession{id: "s1"}, testInstance{id: "i1", name: "order"})

	deleted, ok := ctx.Del("s1", "i1")
	if !ok {
		t.Fatal("expected delete success")
	}
	instance := deleted.(testInstance)
	if instance.name != "order" {
		t.Fatalf("expected order, got %s", instance.name)
	}

	_, ok = ctx.Get("s1", "i1")
	if ok {
		t.Fatal("expected i1 removed")
	}
}

func TestCtxForEachAndList(t *testing.T) {
	ctx := NewBizCtx()
	session := testSession{id: "s1"}
	ctx.Set(session, testInstance{id: "i1", name: "order"})
	ctx.Set(session, testInstance{id: "i2", name: "refund"})

	list := ctx.List("s1")
	if len(list) != 2 {
		t.Fatalf("expected list size 2, got %d", len(list))
	}

	seen := map[string]bool{}
	ctx.ForEach("s1", func(instance BizInstance) {
		seen[instance.BizInstanceId()] = true
	})
	if len(seen) != 2 || !seen["i1"] || !seen["i2"] {
		t.Fatalf("unexpected foreach set: %#v", seen)
	}
}
