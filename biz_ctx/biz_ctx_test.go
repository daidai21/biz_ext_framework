package biz_ctx

import (
	"context"
	"testing"
)

type testInstance struct {
	id   string
	name string
}

func (i testInstance) BizInstanceId() string {
	return i.id
}

var _ BizInstance = testInstance{}
var _ BizSession = (*Session)(nil)

func TestBizSessionOneToMany(t *testing.T) {
	session := NewBizSession("s1")
	session.Set(testInstance{id: "i1", name: "order"})
	session.Set(testInstance{id: "i2", name: "refund"})

	v1, ok := session.Get("i1")
	if !ok {
		t.Fatal("expected i1 exists")
	}
	i1, ok := v1.(testInstance)
	if !ok || i1.name != "order" {
		t.Fatalf("unexpected i1: %#v", v1)
	}

	v2, ok := session.Get("i2")
	if !ok {
		t.Fatal("expected i2 exists")
	}
	i2, ok := v2.(testInstance)
	if !ok || i2.name != "refund" {
		t.Fatalf("unexpected i2: %#v", v2)
	}
}

func TestBizSessionOverwriteAndDel(t *testing.T) {
	session := NewBizSession("s1")
	session.Set(testInstance{id: "i1", name: "old"})
	session.Set(testInstance{id: "i1", name: "new"})

	value, ok := session.Get("i1")
	if !ok {
		t.Fatal("expected i1 exists")
	}
	inst := value.(testInstance)
	if inst.name != "new" {
		t.Fatalf("expected new, got %s", inst.name)
	}

	deleted, ok := session.Del("i1")
	if !ok {
		t.Fatal("expected del success")
	}
	delInst := deleted.(testInstance)
	if delInst.name != "new" {
		t.Fatalf("expected deleted new, got %s", delInst.name)
	}

	_, ok = session.Get("i1")
	if ok {
		t.Fatal("expected i1 removed")
	}
}

func TestBizSessionForEachAndList(t *testing.T) {
	session := NewBizSession("s1")
	session.Set(testInstance{id: "i1", name: "order"})
	session.Set(testInstance{id: "i2", name: "refund"})

	list := session.List()
	if len(list) != 2 {
		t.Fatalf("expected list 2, got %d", len(list))
	}

	seen := map[string]bool{}
	session.ForEach(func(instance BizInstance) {
		seen[instance.BizInstanceId()] = true
	})
	if len(seen) != 2 || !seen["i1"] || !seen["i2"] {
		t.Fatalf("unexpected foreach: %#v", seen)
	}
}

func TestBizSessionContextHelpers(t *testing.T) {
	if _, ok := BizSessionFromContext(context.Background()); ok {
		t.Fatal("expected no session in empty context")
	}

	session := NewBizSession("s1")
	ctx := WithBizSession(context.Background(), session)
	got, ok := BizSessionFromContext(ctx)
	if !ok {
		t.Fatal("expected session in context")
	}
	if got.BizSessionId() != "s1" {
		t.Fatalf("expected s1, got %s", got.BizSessionId())
	}
}
