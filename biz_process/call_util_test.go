package biz_process

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestReqHashStable(t *testing.T) {
	type sampleReq struct {
		OrderID string            `json:"order_id"`
		Ext     map[string]string `json:"ext"`
	}

	req1 := sampleReq{
		OrderID: "1001",
		Ext: map[string]string{
			"b": "2",
			"a": "1",
		},
	}
	req2 := sampleReq{
		OrderID: "1001",
		Ext: map[string]string{
			"a": "1",
			"b": "2",
		},
	}

	hash1, err := ReqHash(req1)
	if err != nil {
		t.Fatalf("req hash failed: %v", err)
	}
	hash2, err := ReqHash(req2)
	if err != nil {
		t.Fatalf("req hash failed: %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("expected stable hash, got %q != %q", hash1, hash2)
	}
}

func TestReqHashUsesTypeName(t *testing.T) {
	type reqA struct {
		Value string `json:"value"`
	}
	type reqB struct {
		Value string `json:"value"`
	}

	hashA, err := ReqHash(reqA{Value: "same"})
	if err != nil {
		t.Fatalf("req hash failed: %v", err)
	}
	hashB, err := ReqHash(reqB{Value: "same"})
	if err != nil {
		t.Fatalf("req hash failed: %v", err)
	}
	if hashA == hashB {
		t.Fatalf("expected different hash for different request types, got %q", hashA)
	}
}

func TestCallWithCacheHitsByRequest(t *testing.T) {
	ctx := WithCallCache(context.Background())

	var called atomic.Int32
	call := func(context.Context, string) (string, error) {
		called.Add(1)
		return "ok", nil
	}

	out1, err := CallWithCache(ctx, "req-1", call)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	out2, err := CallWithCache(ctx, "req-1", call)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if out1 != "ok" || out2 != "ok" {
		t.Fatalf("unexpected outputs: %q, %q", out1, out2)
	}
	if called.Load() != 1 {
		t.Fatalf("expected one upstream call, got %d", called.Load())
	}
}

func TestCallWithCacheBypassesWhenHashFails(t *testing.T) {
	ctx := WithCallCache(context.Background())

	var called atomic.Int32
	call := func(context.Context, func()) (string, error) {
		called.Add(1)
		return "ok", nil
	}

	_, err := CallWithCache(ctx, func() {}, call)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	_, err = CallWithCache(ctx, func() {}, call)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if called.Load() != 2 {
		t.Fatalf("expected direct calls without cache, got %d", called.Load())
	}
}

func TestCallWithCacheAutoCreatesLocalCacheWhenMissing(t *testing.T) {
	var called atomic.Int32
	call := func(context.Context) (string, error) {
		called.Add(1)
		return "ok", nil
	}

	result, err := CallWithCacheKey(context.Background(), "same-key", call)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: %q", result)
	}
	if called.Load() != 1 {
		t.Fatalf("expected one upstream call, got %d", called.Load())
	}
}

func TestCallWithCacheKeyHitsAfterSet(t *testing.T) {
	ctx := WithCallCache(context.Background())

	var called atomic.Int32
	call := func(context.Context) (string, error) {
		called.Add(1)
		return "shared", nil
	}

	result1, err := CallWithCacheKey(ctx, "same-key", call)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	result2, err := CallWithCacheKey(ctx, "same-key", call)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("expected cached hit after first call, got %d upstream calls", called.Load())
	}
	if result1 != "shared" || result2 != "shared" {
		t.Fatalf("unexpected results: %q, %q", result1, result2)
	}
}

func TestCallWithCacheDoesNotCacheError(t *testing.T) {
	ctx := WithCallCache(context.Background())

	var called atomic.Int32
	expectedErr := errors.New("upstream failed")
	call := func(context.Context, string) (string, error) {
		called.Add(1)
		return "", expectedErr
	}

	_, err := CallWithCache(ctx, "req-err", call)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected upstream error, got %v", err)
	}
	_, err = CallWithCache(ctx, "req-err", call)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected upstream error, got %v", err)
	}
	if called.Load() != 2 {
		t.Fatalf("expected error results not to be cached, got %d calls", called.Load())
	}
}

func TestCallWithCacheKeyRequiresCall(t *testing.T) {
	_, err := CallWithCacheKey[string](context.Background(), "k", nil)
	if !errors.Is(err, ErrNilCallFunc) {
		t.Fatalf("expected ErrNilCallFunc, got %v", err)
	}
}
