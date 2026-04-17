package biz_ctx

import (
	"context"
	"testing"
)

func TestSessionZeroValuePaths(t *testing.T) {
	var session Session

	if _, ok := session.Get("missing"); ok {
		t.Fatal("expected empty zero-value session get to miss")
	}
	if _, ok := session.Del("missing"); ok {
		t.Fatal("expected empty zero-value session del to miss")
	}

	called := false
	session.ForEach(func(instance BizInstance) {
		called = true
	})
	if called {
		t.Fatal("expected empty session foreach not to call")
	}
	if list := session.List(); list != nil {
		t.Fatalf("expected nil list for empty session, got %#v", list)
	}

	session.Set(testInstance{id: "i1", name: "order"})
	if value, ok := session.Get("i1"); !ok || value.(testInstance).name != "order" {
		t.Fatalf("unexpected zero-value session set/get result: %v %v", value, ok)
	}
	if _, ok := session.Del("missing"); ok {
		t.Fatal("expected missing delete to remain false")
	}
}

func TestBizSessionFromContextInvalidValues(t *testing.T) {
	if _, ok := BizSessionFromContext(nil); ok {
		t.Fatal("expected nil context lookup to fail")
	}

	ctx := context.WithValue(context.Background(), bizSessionContextKey{}, "bad")
	if _, ok := BizSessionFromContext(ctx); ok {
		t.Fatal("expected invalid context type lookup to fail")
	}

	ctx = WithBizSession(context.Background(), NewBizSession(""))
	if _, ok := BizSessionFromContext(ctx); ok {
		t.Fatal("expected empty session id to fail")
	}
}
