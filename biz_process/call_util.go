package biz_process

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrNilCallFunc           = errors.New("nil call func")
	ErrCallCacheTypeMismatch = errors.New("call cache type mismatch")
)

type callCacheContextKey struct{}

type callCache struct {
	mu      sync.Mutex
	entries map[string]any
}

// WithCallCache stores one call cache container into context.
func WithCallCache(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}

	if cache, ok := callCacheFromContext(ctx); ok && cache != nil {
		return ctx
	}
	return context.WithValue(ctx, callCacheContextKey{}, &callCache{
		entries: make(map[string]any),
	})
}

// ReqHash calculates a stable request hash from request type and JSON payload.
func ReqHash(req any) (string, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	typeName := "<nil>"
	if req != nil {
		typeName = reflect.TypeOf(req).String()
	}

	sum := sha256.Sum256(append([]byte(typeName+":"), payload...))
	return hex.EncodeToString(sum[:]), nil
}

// CallWithCache runs call once per request hash within the same context cache.
// When no cache is attached to ctx, or request hashing fails, it falls back to direct execution.
func CallWithCache[Req any, Resp any](ctx context.Context, req Req, call func(context.Context, Req) (Resp, error)) (Resp, error) {
	if call == nil {
		var zero Resp
		return zero, ErrNilCallFunc
	}

	key, err := ReqHash(req)
	if err != nil {
		return call(ctx, req)
	}

	return CallWithCacheKey(ctx, key, func(ctx context.Context) (Resp, error) {
		return call(ctx, req)
	})
}

// CallWithCacheKey runs call once per cache key within the same context cache.
// Successful results are cached, failed results are not.
func CallWithCacheKey[Resp any](ctx context.Context, key string, call func(context.Context) (Resp, error)) (Resp, error) {
	if call == nil {
		var zero Resp
		return zero, ErrNilCallFunc
	}
	if key == "" {
		return call(ctx)
	}

	cache, ok := callCacheFromContext(ctx)
	if !ok || cache == nil {
		ctx = WithCallCache(ctx)
		cache, _ = callCacheFromContext(ctx)
	}

	if cached, ok := cache.get(key); ok {
		value, ok := cached.(Resp)
		if !ok {
			var zero Resp
			return zero, fmt.Errorf("%w: key=%q cached=%T", ErrCallCacheTypeMismatch, key, cached)
		}
		return value, nil
	}

	value, err := call(ctx)
	if err != nil {
		var zero Resp
		return zero, err
	}

	cache.set(key, value)
	return value, nil
}

func callCacheFromContext(ctx context.Context) (*callCache, bool) {
	if ctx == nil {
		return nil, false
	}
	cache, ok := ctx.Value(callCacheContextKey{}).(*callCache)
	if !ok || cache == nil {
		return nil, false
	}
	return cache, true
}

func (c *callCache) get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.entries[key]
	return value, ok
}

func (c *callCache) set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[string]any)
	}
	c.entries[key] = value
}
